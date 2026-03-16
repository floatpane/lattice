package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
)

// Constructor is the function signature for creating a module instance.
type Constructor func(config.ModuleConfig) module.Module

var (
	mu       sync.RWMutex
	registry = map[string]Constructor{}
)

// Register adds a module constructor to the global registry.
// Typically called from an init() function in the module's package.
//
// Example:
//
//	func init() {
//	    registry.Register("spotify", NewSpotifyModule)
//	}
func Register(name string, ctor Constructor) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("lattice: module %q already registered", name))
	}
	registry[name] = ctor
}

// Get returns the constructor for the given module name, or nil.
func Get(name string) Constructor {
	mu.RLock()
	defer mu.RUnlock()
	return registry[name]
}

// Reset clears the registry. Only for use in tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	registry = map[string]Constructor{}
}

// List returns all registered module names, sorted.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
