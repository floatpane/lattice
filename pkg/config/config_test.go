package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Columns != 2 {
		t.Errorf("expected 2 columns, got %d", cfg.Columns)
	}
	if len(cfg.Modules) == 0 {
		t.Fatal("expected default modules, got none")
	}
	types := make(map[string]bool)
	for _, m := range cfg.Modules {
		types[m.Type] = true
	}
	for _, want := range []string{"greeting", "clock", "system", "github", "weather", "uptime"} {
		if !types[want] {
			t.Errorf("default config missing module %q", want)
		}
	}
}

func TestLoadCreatesFileOnMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := Load()
	if cfg.Columns != 2 {
		t.Errorf("expected default columns, got %d", cfg.Columns)
	}

	path := filepath.Join(tmp, ".config", "lattice", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
}

func TestLoadReadsExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "lattice")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	yaml := `columns: 3
modules:
  - type: clock
  - type: system
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if cfg.Columns != 3 {
		t.Errorf("expected 3 columns, got %d", cfg.Columns)
	}
	if len(cfg.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(cfg.Modules))
	}
}

func TestLoadFixesInvalidColumns(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "lattice")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	yaml := `columns: 0
modules:
  - type: clock
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if cfg.Columns != 2 {
		t.Errorf("expected columns fixed to 2, got %d", cfg.Columns)
	}
}

func TestLoadFallsBackOnEmptyModules(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "lattice")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	yaml := `columns: 2
modules: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if len(cfg.Modules) == 0 {
		t.Error("expected default modules on empty list")
	}
}

func TestLoadHandlesInvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "lattice")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("{{{{not yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if cfg.Columns != 2 {
		t.Errorf("expected default on bad yaml, got columns=%d", cfg.Columns)
	}
}

func TestModuleConfigGet(t *testing.T) {
	mc := ModuleConfig{
		Type:   "test",
		Config: map[string]string{"key": "val"},
	}

	// From config map
	if got := mc.Get("key", "", "default"); got != "val" {
		t.Errorf("expected 'val', got %q", got)
	}

	// From env var
	t.Setenv("TEST_ENV_KEY", "envval")
	if got := mc.Get("missing", "TEST_ENV_KEY", "default"); got != "envval" {
		t.Errorf("expected 'envval', got %q", got)
	}

	// Fallback
	if got := mc.Get("missing", "NONEXISTENT_VAR", "fb"); got != "fb" {
		t.Errorf("expected 'fb', got %q", got)
	}

	// Config takes priority over env
	t.Setenv("TEST_ENV_KEY2", "envval2")
	if got := mc.Get("key", "TEST_ENV_KEY2", "default"); got != "val" {
		t.Errorf("expected config 'val' over env, got %q", got)
	}
}

func TestModuleConfigGetEmptyConfig(t *testing.T) {
	mc := ModuleConfig{Type: "test"}

	if got := mc.Get("anything", "", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}
