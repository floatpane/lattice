package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/floatpane/lattice/internal/layout"
	"github.com/floatpane/lattice/internal/plugin"
	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"

	// Built-in modules register themselves via init().
	_ "github.com/floatpane/lattice/internal/modules"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "import":
			cmdImport(os.Args[2:])
			return
		case "remove":
			cmdRemove(os.Args[2:])
			return
		case "list":
			cmdList()
			return
		case "help", "--help", "-h":
			cmdHelp()
			return
		}
	}

	runDashboard()
}

// --- Dashboard ---

func runDashboard() {
	cfg := config.Load()

	var mods []module.Module
	for _, mc := range cfg.Modules {
		// Try built-in registry first.
		if ctor := registry.Get(mc.Type); ctor != nil {
			mods = append(mods, ctor(mc))
			continue
		}
		// Try external plugin binary (lattice-<name> in PATH or plugins dir).
		if bin := findPlugin(mc.Type); bin != "" {
			mods = append(mods, plugin.NewExternalModule(bin, mc))
			continue
		}
	}

	p := tea.NewProgram(
		&app{modules: mods, columns: cfg.Columns},
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// findPlugin looks for a "lattice-<name>" binary in the plugins dir and PATH.
func findPlugin(name string) string {
	binName := "lattice-" + name

	// 1. Check ~/.config/lattice/plugins/
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".config", "lattice", "plugins", binName)
		if isExecutable(p) {
			return p
		}
	}

	// 2. Check PATH
	if p, err := exec.LookPath(binName); err == nil {
		return p
	}

	return ""
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0111 != 0
}

type app struct {
	modules        []module.Module
	columns        int
	width          int
	height         int
	placements     []layout.ScreenPlacement // image placements from last render
	needsImageDraw bool                     // true when images need to be (re)drawn
}

func (a *app) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, m := range a.modules {
		if cmd := m.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.needsImageDraw = true
		return a, nil
	}

	var cmds []tea.Cmd
	for _, m := range a.modules {
		if cmd := m.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Only re-display images when something changed (new upload, resize, track change).
	if a.needsImageDraw && len(a.placements) > 0 {
		a.needsImageDraw = false
		cmds = append(cmds, renderImagePlacements(a.placements))
	}

	return a, tea.Batch(cmds...)
}

func (a *app) View() tea.View {
	var content string
	if a.width == 0 {
		content = "Starting Lattice…"
	} else {
		var p []layout.ScreenPlacement
		content, p = layout.Render(a.modules, a.columns, a.width, a.height)

		// Check if placements changed (new image, track change, resize, etc.)
		if !samePlacements(a.placements, p) {
			a.needsImageDraw = true
		}
		a.placements = p
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderImagePlacements builds a tea.Raw command that displays images at their
// absolute screen positions using cursor save/restore.
func renderImagePlacements(placements []layout.ScreenPlacement) tea.Cmd {
	var b strings.Builder
	for _, p := range placements {
		// Save cursor, move to absolute position, render image, restore cursor
		fmt.Fprintf(&b, "\x1b[s\x1b[%d;%dH%s\x1b[u", p.Row, p.Col, p.Escape)
	}
	raw := b.String()
	return tea.Raw(raw)
}

func samePlacements(a, b []layout.ScreenPlacement) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Row != b[i].Row || a[i].Col != b[i].Col || a[i].Escape != b[i].Escape {
			return false
		}
	}
	return true
}

// --- CLI subcommands ---

func pluginsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lattice", "plugins")
}

func cmdImport(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: lattice import <go-package>")
		fmt.Println()
		fmt.Println("Installs a plugin binary. The package must produce a binary")
		fmt.Println("named lattice-<name> (e.g., lattice-spotify).")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  lattice import github.com/someone/lattice-spotify@latest")
		os.Exit(1)
	}

	pkg := args[0]
	dir := pluginsDir()

	// Ensure plugins directory exists.
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create plugins dir: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Installing %s...\n", pkg)
	cmd := exec.Command("go", "install", pkg)
	cmd.Env = append(os.Environ(), "GOBIN="+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install: %v\n", err)
		os.Exit(1)
	}

	// Show what was installed.
	entries, _ := os.ReadDir(dir)
	fmt.Println("Installed plugins:")
	for _, e := range entries {
		fmt.Printf("  %s\n", e.Name())
	}
	fmt.Println("\nAdd the module name to your config (~/.config/lattice/config.yaml).")
	fmt.Println("The module name is the binary name minus the 'lattice-' prefix.")
}

func cmdRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: lattice remove <name>")
		fmt.Println("       lattice remove spotify")
		os.Exit(1)
	}

	name := args[0]
	binName := "lattice-" + name
	path := filepath.Join(pluginsDir(), binName)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Plugin %q not found in %s\n", name, pluginsDir())
		} else {
			fmt.Fprintf(os.Stderr, "Failed to remove: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Removed %s\n", name)
	fmt.Println("Don't forget to remove it from your config too.")
}

func cmdList() {
	fmt.Println("Built-in modules:")
	for _, name := range registry.List() {
		fmt.Printf("  %s\n", name)
	}

	dir := pluginsDir()
	entries, _ := os.ReadDir(dir)
	var plugins []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "lattice-") && !e.IsDir() {
			name := strings.TrimPrefix(e.Name(), "lattice-")
			plugins = append(plugins, name)
		}
	}
	if len(plugins) > 0 {
		fmt.Println("\nInstalled plugins:")
		for _, name := range plugins {
			fmt.Printf("  %s\n", name)
		}
	}
}

func cmdHelp() {
	fmt.Println(`Lattice — modular terminal dashboard

Usage:
  lattice              Launch the dashboard
  lattice import <pkg> Install an external plugin module
  lattice remove <name> Remove an installed plugin
  lattice list         Show built-in and installed modules
  lattice help         Show this help

Plugin system:
  Plugins are standalone binaries named "lattice-<name>" that speak
  JSON over stdin/stdout. They are installed to:
    ~/.config/lattice/plugins/

  Install a plugin:
    lattice import github.com/someone/lattice-spotify@latest

  Then add it to your config:
    modules:
      - type: spotify

Creating a plugin:
  A plugin is any binary named lattice-<name> that reads JSON from
  stdin and writes JSON to stdout (one object per line).

  Request types sent by lattice:
    {"type":"init","config":{"key":"val"}}  — once at startup
    {"type":"update"}                        — periodic refresh
    {"type":"view","width":40,"height":10}   — render request

  Response format:
    {"name":"TITLE","content":"rendered text","interval":5}

  The "interval" field (seconds) controls how often "update" is sent.
  See pkg/plugin for the full protocol definition.`)
}
