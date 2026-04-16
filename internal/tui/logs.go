package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
)

const maxLogLines = 5000

type logsBuffer struct {
	mu     sync.Mutex
	buffer []string
	cancel context.CancelFunc
}

func (b *logsBuffer) push(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = append(b.buffer, line)
	if len(b.buffer) > maxLogLines {
		b.buffer = b.buffer[len(b.buffer)-maxLogLines:]
	}
}

func (b *logsBuffer) drain() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buffer) == 0 {
		return nil
	}
	out := b.buffer
	b.buffer = nil
	return out
}

func (b *logsBuffer) stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

type logsModel struct {
	theme    *Theme
	docker   *dockercli.Client
	node     string
	viewport viewport.Model
	lines    []string
	follow   bool
	width    int
	height   int
	buf      *logsBuffer
}

type logLinesMsg struct{ lines []string }
type logsExitMsg struct{}

func newLogs(theme *Theme, docker *dockercli.Client, node string) logsModel {
	vp := viewport.New(80, 20)
	return logsModel{
		theme:    theme,
		docker:   docker,
		node:     node,
		viewport: vp,
		follow:   true,
		buf:      &logsBuffer{},
	}
}

func (m logsModel) Init() tea.Cmd {
	return m.startTailing()
}

func (m logsModel) startTailing() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.buf.cancel = cancel
	node := m.node
	docker := m.docker
	buf := m.buf
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_ = docker.ComposeLogs(ctx, dockercli.ComposeLogOpts{
			Follow: true,
			Node:   node,
			Tail:   500,
			Stdout: pw,
			Stderr: pw,
		})
	}()
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
		for scanner.Scan() {
			buf.push(scanner.Text())
		}
	}()
	return drainCmd(buf)
}

func drainCmd(buf *logsBuffer) tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(_ time.Time) tea.Msg {
		return logLinesMsg{lines: buf.drain()}
	})
}

func (m logsModel) Update(msg tea.Msg) (logsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.buf.stop()
			return m, func() tea.Msg { return logsExitMsg{} }
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
		case "f":
			m.follow = !m.follow
		}
	case logLinesMsg:
		if len(msg.lines) > 0 {
			m.lines = append(m.lines, msg.lines...)
			if len(m.lines) > maxLogLines {
				m.lines = m.lines[len(m.lines)-maxLogLines:]
			}
			m.viewport.SetContent(strings.Join(m.lines, "\n"))
			if m.follow {
				m.viewport.GotoBottom()
			}
		}
		return m, drainCmd(m.buf)
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m logsModel) View() string {
	var title string
	if m.node == "" {
		title = "logs: all services"
	} else {
		title = "logs: " + m.node
	}
	followIndicator := "follow:on"
	if !m.follow {
		followIndicator = "follow:off"
	}
	header := m.theme.Title.Render(title) + m.theme.Dim.Render("  "+followIndicator)
	hint := m.theme.Help.Render("g/G top/bottom · f follow · q back")

	panel := m.theme.Panel.Render(
		fmt.Sprintf("%s\n\n%s\n\n%s", header, m.viewport.View(), hint),
	)
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
