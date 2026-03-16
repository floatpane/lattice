package modules

import (
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	registry.Register("clock", NewClockModule)
}

type ClockModule struct {
	time time.Time
}

type clockTickMsg time.Time

func NewClockModule(_ config.ModuleConfig) module.Module {
	return &ClockModule{time: time.Now()}
}

func (m *ClockModule) Name() string { return "CLOCK" }

func (m *ClockModule) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return clockTickMsg(t) })
}

func (m *ClockModule) Update(msg tea.Msg) tea.Cmd {
	if t, ok := msg.(clockTickMsg); ok {
		m.time = time.Time(t)
		return tea.Tick(time.Second, func(t time.Time) tea.Msg { return clockTickMsg(t) })
	}
	return nil
}

func (m *ClockModule) View(width, height int) string {
	timeStr := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Render(m.time.Format("15:04:05"))

	dateStr := lipgloss.NewStyle().
		Foreground(styles.Subtle).
		Render(m.time.Format("Monday, Jan 02"))

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(width, lipgloss.Center, timeStr),
		lipgloss.PlaceHorizontal(width, lipgloss.Center, dateStr),
	)
}

func (m *ClockModule) MinSize() (int, int) { return 20, 4 }
