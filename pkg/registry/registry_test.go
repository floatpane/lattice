package registry

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
)

type stubModule struct{}

func (s *stubModule) Name() string                  { return "STUB" }
func (s *stubModule) Init() tea.Cmd                 { return nil }
func (s *stubModule) Update(_ tea.Msg) tea.Cmd      { return nil }
func (s *stubModule) View(_, _ int) string           { return "stub" }
func (s *stubModule) MinSize() (int, int)            { return 10, 3 }

func stubCtor(_ config.ModuleConfig) module.Module {
	return &stubModule{}
}

func TestRegisterAndGet(t *testing.T) {
	Reset()
	Register("testmod", stubCtor)

	ctor := Get("testmod")
	if ctor == nil {
		t.Fatal("expected constructor, got nil")
	}

	mod := ctor(config.ModuleConfig{})
	if mod.Name() != "STUB" {
		t.Errorf("expected 'STUB', got %q", mod.Name())
	}
}

func TestGetUnknown(t *testing.T) {
	Reset()
	if Get("nonexistent") != nil {
		t.Error("expected nil for unknown module")
	}
}

func TestList(t *testing.T) {
	Reset()
	Register("beta", stubCtor)
	Register("alpha", stubCtor)
	Register("gamma", stubCtor)

	names := List()
	if len(names) != 3 {
		t.Fatalf("expected 3, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" || names[2] != "gamma" {
		t.Errorf("expected sorted, got %v", names)
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	Reset()
	Register("dup", stubCtor)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate register")
		}
	}()
	Register("dup", stubCtor)
}

func TestReset(t *testing.T) {
	Reset()
	Register("temp", stubCtor)
	if len(List()) != 1 {
		t.Fatal("expected 1 registered")
	}
	Reset()
	if len(List()) != 0 {
		t.Error("expected 0 after reset")
	}
}
