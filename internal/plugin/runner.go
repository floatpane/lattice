package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	proto "github.com/floatpane/lattice/pkg/plugin"
)

// ExternalModule wraps a plugin binary as a module.Module.
type ExternalModule struct {
	binPath  string
	cfg      config.ModuleConfig
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	scanner  *bufio.Scanner
	mu       sync.Mutex
	name     string
	content  string
	minW     int
	minH     int
	interval time.Duration
	started  bool
}

// pluginResponseMsg carries a response from the plugin process.
type pluginResponseMsg struct {
	binPath string
	resp    proto.Response
	err     error
}

// pluginTickMsg triggers an "update" request to the plugin.
type pluginTickMsg struct {
	binPath string
}

// NewExternalModule creates a module backed by an external plugin binary.
func NewExternalModule(binPath string, cfg config.ModuleConfig) module.Module {
	return &ExternalModule{
		binPath: binPath,
		cfg:     cfg,
		name:    cfg.Type,
		minW:    30,
		minH:    4,
		content: "Loading…",
	}
}

func (m *ExternalModule) Name() string { return m.name }

func (m *ExternalModule) Init() tea.Cmd {
	return m.startPlugin
}

func (m *ExternalModule) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case pluginResponseMsg:
		if msg.binPath != m.binPath {
			return nil
		}
		if msg.err != nil {
			m.content = fmt.Sprintf("plugin error: %v", msg.err)
			return nil
		}
		r := msg.resp
		if r.Name != "" {
			m.name = r.Name
		}
		if r.MinWidth > 0 {
			m.minW = r.MinWidth
		}
		if r.MinHeight > 0 {
			m.minH = r.MinHeight
		}
		if r.Interval > 0 {
			m.interval = time.Duration(r.Interval) * time.Second
		}
		if r.Error != "" {
			m.content = r.Error
		} else if r.Content != "" {
			m.content = r.Content
		}

		if m.interval > 0 {
			bp := m.binPath
			return tea.Tick(m.interval, func(_ time.Time) tea.Msg {
				return pluginTickMsg{binPath: bp}
			})
		}
		return nil

	case pluginTickMsg:
		if msg.binPath != m.binPath {
			return nil
		}
		return m.sendUpdate
	}
	return nil
}

func (m *ExternalModule) View(width, height int) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started && m.stdin != nil {
		req := proto.Request{Type: "view", Width: width, Height: height}
		if data, err := json.Marshal(req); err == nil {
			_, _ = m.stdin.Write(append(data, '\n'))
			if m.scanner.Scan() {
				var resp proto.Response
				if err := json.Unmarshal(m.scanner.Bytes(), &resp); err == nil && resp.Content != "" {
					m.content = resp.Content
				}
			}
		}
	}
	return m.content
}

func (m *ExternalModule) MinSize() (int, int) { return m.minW, m.minH }

func (m *ExternalModule) startPlugin() tea.Msg {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cmd = exec.Command(m.binPath)
	stdin, err := m.cmd.StdinPipe()
	if err != nil {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("stdin pipe: %w", err)}
	}
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("stdout pipe: %w", err)}
	}
	m.stdin = stdin
	m.scanner = bufio.NewScanner(stdout)
	m.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if err := m.cmd.Start(); err != nil {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("start: %w", err)}
	}
	m.started = true

	req := proto.Request{Type: "init", Config: m.cfg.Config}
	data, _ := json.Marshal(req)
	_, _ = m.stdin.Write(append(data, '\n'))

	if !m.scanner.Scan() {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("no init response")}
	}

	var resp proto.Response
	if err := json.Unmarshal(m.scanner.Bytes(), &resp); err != nil {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("bad init response: %w", err)}
	}

	return pluginResponseMsg{binPath: m.binPath, resp: resp}
}

func (m *ExternalModule) sendUpdate() tea.Msg {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started || m.stdin == nil {
		return nil
	}

	req := proto.Request{Type: "update"}
	data, _ := json.Marshal(req)
	_, _ = m.stdin.Write(append(data, '\n'))

	if !m.scanner.Scan() {
		return pluginResponseMsg{binPath: m.binPath, err: fmt.Errorf("plugin stopped")}
	}

	var resp proto.Response
	if err := json.Unmarshal(m.scanner.Bytes(), &resp); err != nil {
		return pluginResponseMsg{binPath: m.binPath, err: err}
	}

	return pluginResponseMsg{binPath: m.binPath, resp: resp}
}
