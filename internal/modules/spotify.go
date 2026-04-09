package modules

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/godbus/dbus/v5"

	_ "image/gif"
	_ "image/jpeg"
)

const (
	spotifyArtCols = 10 // cell columns for album art display
	spotifyArtRows = 5  // cell rows for album art display
)

// spotifyNextImageID is an auto-incrementing ID for kitty image uploads.
var spotifyNextImageID uint32 = 900

func init() {
	registry.Register("spotify", NewSpotifyModule)
}

type SpotifyModule struct {
	title     string
	artist    string
	album     string
	artURL    string
	artBase64 string // base64-encoded PNG for upload
	artReady  bool   // true once image has been uploaded to terminal
	imgID     uint32 // kitty image ID for the uploaded art
	position  time.Duration
	duration  time.Duration
	playing   bool
	status    string
	showArt   bool
	kitty     bool
}

type spotifyDataMsg struct {
	title    string
	artist   string
	album    string
	artURL   string
	position time.Duration
	duration time.Duration
	playing  bool
	status   string
}

type spotifyArtMsg struct {
	base64PNG string
}

func NewSpotifyModule(cfg config.ModuleConfig) module.Module {
	showArt := cfg.Get("show_art", "", "true") != "false"
	return &SpotifyModule{
		status:  "Loading…",
		showArt: showArt,
		kitty:   detectKittySupport(),
	}
}

func detectKittySupport() bool {
	term := os.Getenv("TERM")
	termProg := os.Getenv("TERM_PROGRAM")
	if strings.Contains(term, "kitty") || strings.Contains(term, "xterm-kitty") {
		return true
	}
	switch termProg {
	case "ghostty", "WezTerm":
		return true
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" || os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return true
	}
	return false
}

func (m *SpotifyModule) Name() string { return "SPOTIFY" }

func (m *SpotifyModule) Init() tea.Cmd {
	return fetchSpotifyData
}

func (m *SpotifyModule) Update(msg tea.Msg) tea.Cmd {
	switch data := msg.(type) {
	case spotifyArtMsg:
		if data.base64PNG == "" {
			return nil
		}
		m.artBase64 = data.base64PNG

		// Allocate a new image ID and upload via tea.Raw
		spotifyNextImageID++
		m.imgID = spotifyNextImageID
		m.artReady = true

		// Upload (transmit only, a=t) via tea.Raw so it goes directly to the terminal
		return tea.Raw(kittyUploadSequence(m.artBase64, m.imgID))

	case spotifyDataMsg:
		oldArtURL := m.artURL
		oldImgID := m.imgID
		hadArt := m.artReady

		if data.artURL != m.artURL {
			m.artReady = false
			m.artBase64 = ""
			m.imgID = 0
		}

		// If title cleared (nothing playing), delete the old image
		if data.title == "" && hadArt && oldImgID != 0 {
			m.artReady = false
			m.artBase64 = ""
			m.imgID = 0
			m.title = ""
			m.artist = ""
			m.album = ""
			m.artURL = ""
			m.position = 0
			m.duration = 0
			m.playing = false
			m.status = data.status

			return tea.Batch(
				tea.Raw(kittyDeleteSequence(oldImgID)),
				m.scheduleNext(),
			)
		}

		m.title = data.title
		m.artist = data.artist
		m.album = data.album
		m.artURL = data.artURL
		m.position = data.position
		m.duration = data.duration
		m.playing = data.playing
		m.status = data.status

		var cmds []tea.Cmd

		// Delete old image if art URL changed
		if oldArtURL != data.artURL && hadArt && oldImgID != 0 {
			cmds = append(cmds, tea.Raw(kittyDeleteSequence(oldImgID)))
		}

		if m.showArt && m.kitty && m.artURL != "" && !m.artReady {
			artURL := m.artURL
			cmds = append(cmds, func() tea.Msg { return downloadArt(artURL) })
		}

		cmds = append(cmds, m.scheduleNext())
		return tea.Batch(cmds...)
	}
	return nil
}

func (m *SpotifyModule) scheduleNext() tea.Cmd {
	interval := 1 * time.Second
	if !m.playing {
		interval = 5 * time.Second
	}
	return tea.Tick(interval, func(_ time.Time) tea.Msg {
		return fetchSpotifyDataNow()
	})
}

func (m *SpotifyModule) View(width, height int) string {
	if m.title == "" {
		msg := m.status
		if msg == "" {
			msg = "Nothing playing"
		}
		return lipgloss.NewStyle().Foreground(styles.DimText).Render(msg)
	}

	progress := 0.0
	if m.duration > 0 {
		progress = float64(m.position) / float64(m.duration) * 100
	}

	icon := "▶"
	if !m.playing {
		icon = "⏸"
	}

	posStr := formatSpotifyDuration(m.position)
	durStr := formatSpotifyDuration(m.duration)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF"))
	artistStyle := lipgloss.NewStyle().Foreground(styles.DimText)
	timeStyle := lipgloss.NewStyle().Foreground(styles.DimText)

	hasArt := m.showArt && m.kitty && m.artReady
	textWidth := width
	if hasArt {
		textWidth = width - spotifyArtCols - 2 // art columns + gap
	}
	if textWidth < 20 {
		textWidth = 20
	}

	titleText := styles.Truncate(m.title, textWidth)
	var subtitleText string
	if m.album != "" {
		subtitleText = styles.Truncate(m.artist+" — "+m.album, textWidth)
	} else {
		subtitleText = styles.Truncate(m.artist, textWidth)
	}

	barWidth := textWidth - len(posStr) - len(durStr) - len(icon) - 4
	if barWidth < 10 {
		barWidth = 10
	}

	progressLine := fmt.Sprintf("%s %s %s %s",
		icon,
		timeStyle.Render(posStr),
		styles.RenderBar(progress, barWidth, styles.Accent),
		timeStyle.Render(durStr),
	)

	textBlock := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(titleText),
		artistStyle.Render(subtitleText),
		"",
		progressLine,
	)

	if hasArt {
		// Reserve blank space matching the art cell dimensions.
		artSpace := strings.Repeat(" ", spotifyArtCols)
		var artLines []string
		for i := 0; i < spotifyArtRows; i++ {
			artLines = append(artLines, artSpace)
		}
		artBlock := strings.Join(artLines, "\n")

		// Vertically center the text relative to art height.
		textLines := strings.Count(textBlock, "\n") + 1
		if textLines < spotifyArtRows {
			pad := (spotifyArtRows - textLines) / 2
			prefix := strings.Repeat("\n", pad)
			textBlock = prefix + textBlock
		}

		return lipgloss.JoinHorizontal(lipgloss.Top, artBlock, "  ", textBlock)
	}

	return textBlock
}

// ImagePlacements implements module.ImagePlacer.
// Returns the kitty display-by-ID escape positioned at row 0, col 0
// (relative to the module's content area) with cell size constraints.
func (m *SpotifyModule) ImagePlacements() []module.ImagePlacement {
	if !m.showArt || !m.kitty || !m.artReady || m.imgID == 0 {
		return nil
	}
	// a=p: display previously uploaded image by ID
	// c/r: constrain to this many cell columns/rows
	// C=1: don't move cursor
	// q=2: quiet
	escape := fmt.Sprintf("\x1b_Ga=p,i=%d,c=%d,r=%d,q=2,C=1\x1b\\",
		m.imgID, spotifyArtCols, spotifyArtRows)
	return []module.ImagePlacement{
		{
			Row:    0,
			Col:    0,
			Escape: escape,
		},
	}
}

func (m *SpotifyModule) MinSize() (int, int) { return 40, 6 }

func fetchSpotifyData() tea.Msg {
	return fetchSpotifyDataNow()
}

func fetchSpotifyDataNow() spotifyDataMsg {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return spotifyDataMsg{status: "D-Bus unavailable"}
	}
	defer func() { _ = conn.Close() }()

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")

	statusVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		return spotifyDataMsg{status: "Spotify not running"}
	}
	playbackStatus, _ := statusVariant.Value().(string)

	metaVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return spotifyDataMsg{status: "No metadata"}
	}
	metadata, _ := metaVariant.Value().(map[string]dbus.Variant)

	title := variantString(metadata, "xesam:title")
	artists := variantStringSlice(metadata, "xesam:artist")
	album := variantString(metadata, "xesam:album")
	artURL := variantString(metadata, "mpris:artUrl")
	lengthUs := variantInt64(metadata, "mpris:length")

	posVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
	positionUs := int64(0)
	if err == nil {
		positionUs, _ = posVariant.Value().(int64)
	}

	artist := strings.Join(artists, ", ")

	return spotifyDataMsg{
		title:    title,
		artist:   artist,
		album:    album,
		artURL:   artURL,
		position: time.Duration(positionUs) * time.Microsecond,
		duration: time.Duration(lengthUs) * time.Microsecond,
		playing:  playbackStatus == "Playing",
	}
}

// downloadArt fetches the album art and re-encodes it as base64 PNG.
func downloadArt(url string) spotifyArtMsg {
	if url == "" {
		return spotifyArtMsg{}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return spotifyArtMsg{}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return spotifyArtMsg{}
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return spotifyArtMsg{}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return spotifyArtMsg{}
	}

	resized := resizeImage(img, 256, 256)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return spotifyArtMsg{}
	}

	return spotifyArtMsg{
		base64PNG: base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
}

// kittyUploadSequence builds the escape sequence to upload (transmit only, a=t)
// a PNG image to the terminal's memory with the given ID.
func kittyUploadSequence(payload string, id uint32) string {
	if payload == "" {
		return ""
	}

	const chunkSize = 4096
	var b strings.Builder

	for offset := 0; offset < len(payload); offset += chunkSize {
		end := offset + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		more := "0"
		if end < len(payload) {
			more = "1"
		}

		chunk := payload[offset:end]
		if offset == 0 {
			fmt.Fprintf(&b, "\x1b_Gf=100,a=t,i=%d,q=2,m=%s;%s\x1b\\", id, more, chunk)
		} else {
			fmt.Fprintf(&b, "\x1b_Gm=%s;%s\x1b\\", more, chunk)
		}
	}

	return b.String()
}

// kittyDeleteSequence builds the escape sequence to delete an image by ID.
func kittyDeleteSequence(id uint32) string {
	return fmt.Sprintf("\x1b_Ga=d,d=I,i=%d,q=2;\x1b\\", id)
}

// resizeImage scales an image to fit within maxW x maxH using nearest-neighbor.
func resizeImage(src image.Image, maxW, maxH int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	scale := float64(maxW) / float64(srcW)
	if s := float64(maxH) / float64(srcH); s < scale {
		scale = s
	}

	dstW := int(float64(srcW) * scale)
	dstH := int(float64(srcH) * scale)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := bounds.Min.X + x*srcW/dstW
			srcY := bounds.Min.Y + y*srcH/dstH
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func variantString(m map[string]dbus.Variant, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.Value().(string)
	return s
}

func variantStringSlice(m map[string]dbus.Variant, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	ss, _ := v.Value().([]string)
	return ss
}

func variantInt64(m map[string]dbus.Variant, key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.Value().(type) {
	case int64:
		return val
	case uint64:
		return int64(val)
	case int32:
		return int64(val)
	case uint32:
		return int64(val)
	default:
		return 0
	}
}

func formatSpotifyDuration(d time.Duration) string {
	total := int(d.Seconds())
	if total < 0 {
		total = 0
	}
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
