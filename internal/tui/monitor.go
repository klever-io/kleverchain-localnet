package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
)

type monitorModel struct {
	theme   *Theme
	docker  *dockercli.Client
	refresh time.Duration
	noStats bool
	width   int
	height  int
	rows    []monitorRow
	cursor  int
	lastErr error
	updated time.Time
}

type monitorRow struct {
	Service string
	State   string
	Health  string
	CPU     float64
	Mem     string
	Ports   string
}

type monitorTickMsg struct{}
type monitorDataMsg struct {
	rows []monitorRow
	err  error
}

func newMonitor(theme *Theme, docker *dockercli.Client, refresh time.Duration, noStats bool) monitorModel {
	if refresh == 0 {
		refresh = 2 * time.Second
	}
	return monitorModel{
		theme:   theme,
		docker:  docker,
		refresh: refresh,
		noStats: noStats,
	}
}

func (m monitorModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetch(),
		m.tick(),
	)
}

func (m monitorModel) tick() tea.Cmd {
	return tea.Tick(m.refresh, func(time.Time) tea.Msg { return monitorTickMsg{} })
}

func (m monitorModel) fetch() tea.Cmd {
	docker := m.docker
	noStats := m.noStats
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		services, err := docker.ComposePS(ctx)
		if err != nil {
			return monitorDataMsg{err: err}
		}
		statsByName := map[string]dockercli.Stat{}
		if !noStats {
			stats, serr := docker.Stats(ctx)
			if serr == nil {
				for _, s := range stats {
					statsByName[s.Name] = s
				}
			}
		}
		rows := make([]monitorRow, 0, len(services))
		for _, s := range services {
			row := monitorRow{
				Service: s.ServiceName,
				State:   s.State,
				Health:  s.Health,
				Ports:   joinPortsForMonitor(s),
			}
			if st, ok := statsByName[s.Name]; ok {
				row.CPU = st.CPUPct
				row.Mem = st.MemUsage
			}
			rows = append(rows, row)
		}
		return monitorDataMsg{rows: rows}
	}
}

func joinPortsForMonitor(s dockercli.Service) string {
	if len(s.Ports) == 0 {
		return "—"
	}
	out := make([]string, 0, len(s.Ports))
	seen := make(map[string]struct{})
	for _, p := range s.Ports {
		if p.Host == "" {
			continue
		}
		if _, ok := seen[p.Host]; ok {
			continue
		}
		seen[p.Host] = struct{}{}
		out = append(out, p.Host)
	}
	if len(out) == 0 {
		return "—"
	}
	return strings.Join(out, ",")
}

type monitorExitMsg struct{}

func (m monitorModel) Update(msg tea.Msg) (monitorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return monitorExitMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		}
	case monitorTickMsg:
		return m, tea.Batch(m.fetch(), m.tick())
	case monitorDataMsg:
		m.rows = msg.rows
		m.lastErr = msg.err
		m.updated = time.Now()
	}
	return m, nil
}

func (m monitorModel) View() string {
	title := m.theme.Title.Render(fmt.Sprintf("KLOCAL Monitor — auto-refresh %s", m.refresh))
	hint := m.theme.Help.Render("↑/↓ select · q back")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	header := m.theme.TableHeader.Render(fmt.Sprintf("%-14s %-10s %-10s %-7s %-12s %-8s", "SERVICE", "STATE", "HEALTH", "CPU", "MEM", "PORTS"))
	lines = append(lines, header)

	if m.lastErr != nil {
		lines = append(lines, m.theme.ErrStyle.Render("error: "+m.lastErr.Error()))
	}
	if len(m.rows) == 0 && m.lastErr == nil {
		lines = append(lines, m.theme.Dim.Render("no containers running"))
	}

	for i, r := range m.rows {
		cpuStr := fmt.Sprintf("%4.1f%%", r.CPU)
		mem := r.Mem
		if mem == "" {
			mem = "—"
		}
		line := fmt.Sprintf("%-14s %-10s %-10s %-7s %-12s %-8s", r.Service, r.State, healthOrDash(r.Health), cpuStr, mem, r.Ports)
		colored := colorStateStyle(m.theme, r.State, r.Health).Render(line)
		if i == m.cursor {
			colored = m.theme.Selected.Render("▸ ") + colored
		} else {
			colored = "  " + colored
		}
		lines = append(lines, colored)
	}

	lines = append(lines, "")
	if !m.updated.IsZero() {
		lines = append(lines, m.theme.Dim.Render("last refresh: "+m.updated.Format("15:04:05")))
	}
	lines = append(lines, hint)

	panel := m.theme.Panel.Render(strings.Join(lines, "\n"))
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func healthOrDash(h string) string {
	if h == "" {
		return "—"
	}
	return h
}

func colorStateStyle(th *Theme, state, health string) lipgloss.Style {
	switch {
	case state == "running" && (health == "healthy" || health == ""):
		return th.Success
	case state == "running" && health == "starting":
		return th.Warn
	case state == "restarting":
		return th.Warn
	case state == "exited" || health == "unhealthy":
		return th.ErrStyle
	}
	return lipgloss.NewStyle()
}
