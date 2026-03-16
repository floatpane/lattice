package modules

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	registry.Register("weather", NewWeatherModule)
}

type WeatherModule struct {
	city    string
	display string
}

type weatherDataMsg string

func NewWeatherModule(cfg config.ModuleConfig) module.Module {
	return &WeatherModule{
		city:    cfg.Get("city", "LATTICE_CITY", ""),
		display: "Loading…",
	}
}

func (m *WeatherModule) Name() string { return "WEATHER" }

func (m *WeatherModule) Init() tea.Cmd {
	city := m.city
	return func() tea.Msg { return fetchWeather(city) }
}

func (m *WeatherModule) Update(msg tea.Msg) tea.Cmd {
	if data, ok := msg.(weatherDataMsg); ok {
		m.display = string(data)
		city := m.city
		return tea.Tick(15*time.Minute, func(_ time.Time) tea.Msg {
			return fetchWeather(city)
		})
	}
	return nil
}

func (m *WeatherModule) View(width, height int) string {
	return m.display
}

func (m *WeatherModule) MinSize() (int, int) { return 30, 5 }

func fetchWeather(city string) weatherDataMsg {
	url := "https://wttr.in/" + city + "?format=%c+%C\\n🌡+%t+(feels+%f)\\n💨+%w\\n💧+%h"

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "lattice-terminal")

	resp, err := client.Do(req)
	if err != nil {
		return weatherDataMsg(lipgloss.NewStyle().Foreground(styles.DimText).Render("Could not fetch weather"))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return weatherDataMsg(lipgloss.NewStyle().Foreground(styles.DimText).Render("Read error"))
	}

	result := strings.TrimSpace(string(body))
	if strings.Contains(result, "Unknown location") {
		return weatherDataMsg(lipgloss.NewStyle().Foreground(styles.DimText).Render("Set city in config or LATTICE_CITY"))
	}

	return weatherDataMsg(result)
}
