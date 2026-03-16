// Package plugin provides types and a helper for building Lattice plugins.
//
// A plugin is a standalone binary that reads JSON requests from stdin
// and writes JSON responses to stdout, one per line.
//
// Quick start:
//
//	func main() {
//	    plugin.Run(func(req plugin.Request) plugin.Response {
//	        switch req.Type {
//	        case "init":
//	            return plugin.Response{Name: "MY MODULE", Interval: 10}
//	        case "update":
//	            data := fetchSomething()
//	            return plugin.Response{Content: data}
//	        case "view":
//	            return plugin.Response{Content: renderView(req.Width, req.Height)}
//	        }
//	        return plugin.Response{}
//	    })
//	}
package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Request is sent from Lattice to the plugin over stdin.
type Request struct {
	Type   string            `json:"type"`             // "init", "update", "view"
	Config map[string]string `json:"config,omitempty"` // sent with "init"
	Width  int               `json:"width,omitempty"`  // sent with "view"
	Height int               `json:"height,omitempty"` // sent with "view"
}

// Response is sent from the plugin back to Lattice over stdout.
type Response struct {
	// Name is the display title (only needed in "init" response).
	Name string `json:"name,omitempty"`

	// Content is the rendered text for "view" and "update" responses.
	Content string `json:"content,omitempty"`

	// MinWidth and MinHeight are size hints (only needed in "init" response).
	MinWidth  int `json:"min_width,omitempty"`
	MinHeight int `json:"min_height,omitempty"`

	// Interval is how often (in seconds) Lattice should send "update".
	// Set in "init" response. 0 means no periodic updates.
	Interval int `json:"interval,omitempty"`

	// Error, if set, is displayed instead of content.
	Error string `json:"error,omitempty"`
}

// Run starts the plugin loop. It reads requests from stdin and calls
// handler for each one, writing the response to stdout.
// This function never returns under normal operation.
func Run(handler func(Request) Response) {
	RunWith(os.Stdin, os.Stdout, handler)
}

// RunWith is like Run but reads from r and writes to w.
// Useful for testing plugins without real stdin/stdout.
func RunWith(r io.Reader, w io.Writer, handler func(Request) Response) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			enc.Encode(Response{Error: fmt.Sprintf("bad request: %v", err)})
			continue
		}
		resp := handler(req)
		enc.Encode(resp)
	}
}
