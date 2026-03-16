package module

import tea "charm.land/bubbletea/v2"

// Module is the interface every Lattice module must implement.
// Modules are self-contained: they manage their own state, fetching,
// and rendering. The framework handles layout and lifecycle.
//
// External modules should implement this interface and register
// themselves in their init() function using registry.Register().
type Module interface {
	// Name returns the display name shown in the module's title bar.
	Name() string

	// Init returns the initial command (e.g. first data fetch).
	Init() tea.Cmd

	// Update handles messages. Return nil cmd if the message isn't relevant.
	Update(msg tea.Msg) tea.Cmd

	// View renders the module content (without the border/title — the
	// framework wraps it). width and height are the available content area.
	View(width, height int) string

	// MinSize returns the preferred minimum width and height (content area,
	// excluding border/padding). The layout engine uses these as hints.
	MinSize() (width, height int)
}
