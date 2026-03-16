package styles

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"", 5, ""},
		{"ab", 1, "…"},
	}
	for _, tt := range tests {
		got := Truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestTruncateUnicode(t *testing.T) {
	got := Truncate("日本語テスト", 4)
	runes := []rune(got)
	if len(runes) != 4 {
		t.Errorf("expected 4 runes, got %d: %q", len(runes), got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected trailing ellipsis, got %q", got)
	}
}

func TestRenderBarBounds(t *testing.T) {
	// 0%
	bar := RenderBar(0, 10, Accent)
	if bar == "" {
		t.Error("expected non-empty bar at 0%")
	}

	// 100%
	bar = RenderBar(100, 10, Accent)
	if bar == "" {
		t.Error("expected non-empty bar at 100%")
	}

	// Negative clamps to 0
	bar = RenderBar(-50, 10, Accent)
	if bar == "" {
		t.Error("expected non-empty bar at -50%")
	}

	// Over 100 clamps to 100
	bar = RenderBar(200, 10, Accent)
	if bar == "" {
		t.Error("expected non-empty bar at 200%")
	}
}

func TestRenderBarWidth(t *testing.T) {
	// At 50% with width 10, should have roughly 5 filled + 5 empty blocks
	bar := RenderBar(50, 10, lipgloss.Color("#FFF"))
	if len(bar) == 0 {
		t.Error("bar should not be empty")
	}
}

func TestRenderStat(t *testing.T) {
	result := RenderStat("CPU", "45%")
	if result == "" {
		t.Error("expected non-empty stat")
	}
	// Should contain the value somewhere in the output
	if !strings.Contains(result, "45%") {
		t.Errorf("expected stat to contain '45%%', got %q", result)
	}
}

func TestColorsAreDefined(t *testing.T) {
	// Verify all exported colors have both light and dark values
	colors := []compat.AdaptiveColor{Subtle, Accent, Warn, Highlight, DimText}
	for i, c := range colors {
		if c.Light == nil || c.Dark == nil {
			t.Errorf("color %d has nil light or dark value", i)
		}
	}
}
