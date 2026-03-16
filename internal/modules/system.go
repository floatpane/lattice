package modules

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/floatpane/lattice/pkg/config"
	"github.com/floatpane/lattice/pkg/module"
	"github.com/floatpane/lattice/pkg/registry"
	"github.com/floatpane/lattice/pkg/styles"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func init() {
	registry.Register("system", NewSystemModule)
}

type SystemModule struct {
	cpuLoad float64
	memLoad float64
	gpuLoad float64
}

type systemDataMsg struct {
	cpuLoad float64
	memLoad float64
	gpuLoad float64
}

func NewSystemModule(_ config.ModuleConfig) module.Module {
	return &SystemModule{gpuLoad: -1}
}

func (m *SystemModule) Name() string { return "SYSTEM" }

func (m *SystemModule) Init() tea.Cmd {
	return fetchSystemData
}

func (m *SystemModule) Update(msg tea.Msg) tea.Cmd {
	if data, ok := msg.(systemDataMsg); ok {
		m.cpuLoad = data.cpuLoad
		m.memLoad = data.memLoad
		m.gpuLoad = data.gpuLoad
		return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return fetchSystemDataNow() })
	}
	return nil
}

func (m *SystemModule) View(width, height int) string {
	barW := width - 10
	if barW < 10 {
		barW = 10
	}

	lines := fmt.Sprintf("CPU %3.0f%% %s\nMEM %3.0f%% %s",
		m.cpuLoad, styles.RenderBar(m.cpuLoad, barW, styles.Warn),
		m.memLoad, styles.RenderBar(m.memLoad, barW, styles.Highlight),
	)
	if m.gpuLoad >= 0 {
		lines += fmt.Sprintf("\nGPU %3.0f%% %s", m.gpuLoad, styles.RenderBar(m.gpuLoad, barW, styles.Accent))
	}
	return lines
}

func (m *SystemModule) MinSize() (int, int) { return 30, 5 }

func fetchSystemData() tea.Msg {
	return fetchSystemDataNow()
}

func fetchSystemDataNow() systemDataMsg {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(200*time.Millisecond, false)
	cpuVal := 0.0
	if len(c) > 0 {
		cpuVal = c[0]
	}

	gpuVal := -1.0
	out, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits").Output()
	if err == nil {
		s := strings.TrimSpace(string(out))
		if val, err := strconv.ParseFloat(s, 64); err == nil {
			gpuVal = val
		}
	}

	return systemDataMsg{
		cpuLoad: cpuVal,
		memLoad: v.UsedPercent,
		gpuLoad: gpuVal,
	}
}
