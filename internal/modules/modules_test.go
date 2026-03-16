package modules

import (
	"testing"
	"time"

	"github.com/floatpane/lattice/pkg/config"
)

func TestGreetingForTime(t *testing.T) {
	tests := []struct {
		hour int
		want string
	}{
		{0, "Good night"},
		{3, "Good night"},
		{5, "Good night"},
		{6, "Good morning"},
		{11, "Good morning"},
		{12, "Good afternoon"},
		{16, "Good afternoon"},
		{17, "Good evening"},
		{23, "Good evening"},
	}
	for _, tt := range tests {
		tm := time.Date(2024, 1, 1, tt.hour, 0, 0, 0, time.UTC)
		got := greetingForTime(tm)
		if got != tt.want {
			t.Errorf("greetingForTime(hour=%d) = %q, want %q", tt.hour, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{5 * time.Minute, "5m"},
		{90 * time.Minute, "1h 30m"},
		{25 * time.Hour, "1d 1h 0m"},
		{49*time.Hour + 30*time.Minute, "2d 1h 30m"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestClockModuleInterface(t *testing.T) {
	mod := NewClockModule(config.ModuleConfig{})
	if mod.Name() != "CLOCK" {
		t.Errorf("expected 'CLOCK', got %q", mod.Name())
	}
	w, h := mod.MinSize()
	if w == 0 || h == 0 {
		t.Error("expected non-zero min size")
	}
	view := mod.View(40, 10)
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestGreetingModuleInterface(t *testing.T) {
	mod := NewGreetingModule(config.ModuleConfig{
		Config: map[string]string{"name": "TestUser"},
	})
	if mod.Name() != "LATTICE" {
		t.Errorf("expected 'LATTICE', got %q", mod.Name())
	}
	view := mod.View(40, 10)
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestSystemModuleInterface(t *testing.T) {
	mod := NewSystemModule(config.ModuleConfig{})
	if mod.Name() != "SYSTEM" {
		t.Errorf("expected 'SYSTEM', got %q", mod.Name())
	}
	w, h := mod.MinSize()
	if w == 0 || h == 0 {
		t.Error("expected non-zero min size")
	}
}

func TestWeatherModuleInterface(t *testing.T) {
	mod := NewWeatherModule(config.ModuleConfig{})
	if mod.Name() != "WEATHER" {
		t.Errorf("expected 'WEATHER', got %q", mod.Name())
	}
}

func TestUptimeModuleInterface(t *testing.T) {
	mod := NewUptimeModule(config.ModuleConfig{})
	if mod.Name() != "UPTIME" {
		t.Errorf("expected 'UPTIME', got %q", mod.Name())
	}
}

func TestGitHubModuleInterface(t *testing.T) {
	mod := NewGitHubModule(config.ModuleConfig{})
	if mod.Name() != "GITHUB" {
		t.Errorf("expected 'GITHUB', got %q", mod.Name())
	}
	// Without credentials, should show a message
	view := mod.View(40, 10)
	if view == "" {
		t.Error("expected non-empty view even without credentials")
	}
}

func TestGitHubModuleWithUsername(t *testing.T) {
	mod := NewGitHubModule(config.ModuleConfig{
		Config: map[string]string{"username": "testuser"},
	})
	// Should show loading state
	view := mod.View(40, 10)
	if view == "" {
		t.Error("expected non-empty view")
	}
}
