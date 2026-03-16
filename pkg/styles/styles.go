package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

var (
	// Colors — exported for use by modules and layout.
	Subtle    = compat.AdaptiveColor{Light: lipgloss.Color("#D9DCCF"), Dark: lipgloss.Color("#383838")}
	Accent    = compat.AdaptiveColor{Light: lipgloss.Color("#43BF6D"), Dark: lipgloss.Color("#73F59F")}
	Warn      = compat.AdaptiveColor{Light: lipgloss.Color("#F25D94"), Dark: lipgloss.Color("#F55385")}
	Highlight = compat.AdaptiveColor{Light: lipgloss.Color("#874BFD"), Dark: lipgloss.Color("#7D56F4")}
	DimText   = compat.AdaptiveColor{Light: lipgloss.Color("#9B9B9B"), Dark: lipgloss.Color("#5C5C5C")}

	statLabelStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Transform(strings.ToUpper).
			Width(16)

	statValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Bold(true)
)

// RenderBar draws a horizontal bar chart.
func RenderBar(percent float64, width int, c color.Color) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(float64(width) * (percent / 100))
	empty := width - filled
	if empty < 0 {
		empty = 0
	}
	if filled > width {
		filled = width
	}
	return lipgloss.NewStyle().Foreground(c).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(Subtle).Render(strings.Repeat("░", empty))
}

// RenderStat renders a label-value pair.
func RenderStat(label, val string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		statLabelStyle.Render(label),
		statValueStyle.Render(val),
	)
}

// Truncate shortens a string to max runes.
func Truncate(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max-1]) + "…"
	}
	return s
}
