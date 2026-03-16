package layout

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/floatpane/lattice/pkg/module"
)

type mockModule struct {
	name    string
	content string
	w, h    int
}

func (m *mockModule) Name() string                  { return m.name }
func (m *mockModule) Init() tea.Cmd                 { return nil }
func (m *mockModule) Update(_ tea.Msg) tea.Cmd      { return nil }
func (m *mockModule) View(_, _ int) string           { return m.content }
func (m *mockModule) MinSize() (int, int)            { return m.w, m.h }

func TestRenderEmpty(t *testing.T) {
	result := Render(nil, 2, 80, 24)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderSingleModule(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "TEST", content: "hello", w: 20, h: 3},
	}
	result := Render(mods, 2, 80, 24)
	if result == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(result, "TEST") {
		t.Error("expected module title in output")
	}
	if !strings.Contains(result, "hello") {
		t.Error("expected module content in output")
	}
}

func TestRenderMultipleModules(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "A", content: "aaa", w: 20, h: 3},
		&mockModule{name: "B", content: "bbb", w: 20, h: 3},
		&mockModule{name: "C", content: "ccc", w: 20, h: 3},
		&mockModule{name: "D", content: "ddd", w: 20, h: 3},
	}
	result := Render(mods, 2, 100, 40)
	for _, name := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(result, name) {
			t.Errorf("expected %q in output", name)
		}
	}
}

func TestRenderClampsColumns(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "ONLY", content: "one", w: 20, h: 3},
	}
	// 5 columns but only 1 module — should clamp to 1 column
	result := Render(mods, 5, 80, 24)
	if !strings.Contains(result, "ONLY") {
		t.Error("expected module in output even with excess columns")
	}
}

func TestRenderMinColumnWidth(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "NARROW", content: "x", w: 20, h: 3},
	}
	// Very narrow terminal — should still render
	result := Render(mods, 1, 10, 24)
	if result == "" {
		t.Error("expected output even with tiny terminal")
	}
}

func TestRenderZeroColumns(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "A", content: "a", w: 20, h: 3},
	}
	// 0 columns should be treated as 1
	result := Render(mods, 0, 80, 24)
	if !strings.Contains(result, "A") {
		t.Error("expected module with 0 columns (should default to 1)")
	}
}

func TestRenderMinHeight(t *testing.T) {
	mods := []module.Module{
		&mockModule{name: "TINY", content: "x", w: 20, h: 1}, // below minimum of 3
	}
	result := Render(mods, 1, 80, 24)
	if !strings.Contains(result, "TINY") {
		t.Error("expected module with small min height")
	}
}
