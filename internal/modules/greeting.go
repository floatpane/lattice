package modules

import (
	"os"
	"os/user"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	registry.Register("greeting", NewGreetingModule)
}

type GreetingModule struct {
	name string
}

func NewGreetingModule(cfg config.ModuleConfig) module.Module {
	name := cfg.Get("name", "LATTICE_NAME", "")
	if name == "" {
		name = os.Getenv("USER")
	}
	if name == "" {
		if u, err := user.Current(); err == nil {
			name = u.Username
		}
	}
	return &GreetingModule{name: name}
}

func (m *GreetingModule) Name() string { return "LATTICE" }

func (m *GreetingModule) Init() tea.Cmd { return nil }

func (m *GreetingModule) Update(_ tea.Msg) tea.Cmd { return nil }

func (m *GreetingModule) View(width, height int) string {
	greeting := greetingForTime(time.Now())

	greetStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF")).Bold(true)

	line := greetStyle.Render(greeting+", ") + nameStyle.Render(m.name)
	sub := lipgloss.NewStyle().Foreground(styles.DimText).Render("Your terminal dashboard")

	return lipgloss.JoinVertical(lipgloss.Left, line, sub)
}

func (m *GreetingModule) MinSize() (int, int) { return 28, 4 }

func greetingForTime(t time.Time) string {
	h := t.Hour()
	switch {
	case h < 6:
		return "Good night"
	case h < 12:
		return "Good morning"
	case h < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}
