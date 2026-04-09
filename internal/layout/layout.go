package layout

import (
	"strings"

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

// ScreenPlacement is an image placement with absolute screen coordinates.
type ScreenPlacement struct {
	Row    int    // absolute terminal row (1-based)
	Col    int    // absolute terminal col (1-based)
	Escape string // kitty graphics escape sequence
}

// Render arranges modules into a columnar grid.
// It returns the rendered text and any image placements with absolute positions.
func Render(modules []module.Module, columns, termWidth, termHeight int) (string, []ScreenPlacement) {
	if len(modules) == 0 {
		return "", nil
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

	// Track per-column cumulative height (in rendered lines) for positioning.
	colHeights := make([]int, columns)

	// Distribute modules across columns round-robin
	cols := make([][]string, columns)
	var placements []ScreenPlacement

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

		if placer, ok := mod.(module.ImagePlacer); ok {
			for _, ip := range placer.ImagePlacements() {
				// To find the content area within the box, we need to account for:
				// - docStyle Padding(1, 2): 1 row top, 2 cols left
				// - box border top: 1 row
				// - box padding top: 1 row
				// - title line: 1 row
				// - title MarginBottom(1): 1 row
				// - box border left: 1 col, box padding left: 2 cols
				// Total row offset from box top to content: 4
				// Total col offset from box left to content: 3
				absRow := 1 + colHeights[col] + 4 + ip.Row + 1  // +1 for docPadTop, +1 for 1-based
				absCol := 2 + (col * colWidth) + 3 + ip.Col + 1 // +2 for docPadLeft, +1 for 1-based
				placements = append(placements, ScreenPlacement{
					Row:    absRow,
					Col:    absCol,
					Escape: ip.Escape,
				})
			}
		}

		// Count actual rendered lines in the box for accurate height tracking
		boxHeight := strings.Count(box, "\n") + 1
		colHeights[col] += boxHeight

		cols[col] = append(cols[col], box)
	}

	// Join each column vertically, then join columns horizontally
	rendered := make([]string, columns)
	for i, col := range cols {
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, col...)
	}

	text := docStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, rendered...),
	)

	return text, placements
}
