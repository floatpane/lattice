package layout

import (
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/styles"

	"charm.land/lipgloss/v2"
)

var (
	docStyle = lipgloss.NewStyle().Padding(1, 2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Subtle).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(styles.Subtle).
			Bold(true).
			MarginBottom(1)
)

// Render arranges modules into a columnar grid.
func Render(modules []module.Module, columns, termWidth, termHeight int) string {
	if len(modules) == 0 {
		return ""
	}
	if columns < 1 {
		columns = 1
	}
	if columns > len(modules) {
		columns = len(modules)
	}

	// Available width per column (accounting for doc padding of 2 on each side)
	usable := termWidth - 4
	colWidth := usable / columns
	if colWidth < 20 {
		colWidth = 20
	}

	// Content width inside the box (subtract border=2 + padding=4)
	contentWidth := colWidth - 6

	// Distribute modules across columns round-robin
	cols := make([][]string, columns)
	for i, mod := range modules {
		col := i % columns
		_, minH := mod.MinSize()
		if minH < 3 {
			minH = 3
		}

		content := mod.View(contentWidth, minH)
		title := titleStyle.Render(mod.Name())
		inner := lipgloss.JoinVertical(lipgloss.Left, title, content)

		box := boxStyle.Width(colWidth - 2).Height(minH + 2).Render(inner)
		cols[col] = append(cols[col], box)
	}

	// Join each column vertically, then join columns horizontally
	rendered := make([]string, columns)
	for i, col := range cols {
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, col...)
	}

	return docStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, rendered...),
	)
}
