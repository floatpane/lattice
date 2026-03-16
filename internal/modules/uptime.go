package modules

import (
	"fmt"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shirou/gopsutil/v3/host"
)

func init() {
	registry.Register("uptime", NewUptimeModule)
}

type UptimeModule struct {
	bootTime time.Time
	uptime   string
}

type uptimeTickMsg struct{}

func NewUptimeModule(_ config.ModuleConfig) module.Module {
	return &UptimeModule{}
}

func (m *UptimeModule) Name() string { return "UPTIME" }

func (m *UptimeModule) Init() tea.Cmd {
	bt, _ := host.BootTime()
	m.bootTime = time.Unix(int64(bt), 0)
	m.uptime = formatDuration(time.Since(m.bootTime))
	return tea.Tick(time.Minute, func(_ time.Time) tea.Msg { return uptimeTickMsg{} })
}

func (m *UptimeModule) Update(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(uptimeTickMsg); ok {
		m.uptime = formatDuration(time.Since(m.bootTime))
		return tea.Tick(time.Minute, func(_ time.Time) tea.Msg { return uptimeTickMsg{} })
	}
	return nil
}

func (m *UptimeModule) View(width, height int) string {
	upStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	sinceStyle := lipgloss.NewStyle().Foreground(styles.DimText)

	return lipgloss.JoinVertical(lipgloss.Left,
		upStyle.Render(m.uptime),
		sinceStyle.Render("since "+m.bootTime.Format("Jan 02, 15:04")),
	)
}

func (m *UptimeModule) MinSize() (int, int) { return 20, 4 }

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
