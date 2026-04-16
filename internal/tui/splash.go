package tui

import (
	_ "embed"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/klever-io/kleverchain-localnet/internal/version"
)

//go:embed assets/klocal-banner.txt
var bannerSource string

const splashHoldDuration = 1500 * time.Millisecond

type splashTickMsg struct{}
type splashDoneMsg struct{}

type splashModel struct {
	theme       *Theme
	banner      []string
	revealed    int
	totalChars  int
	width       int
	height      int
	done        bool
	holding     bool
	fastForward bool
}

func newSplash(theme *Theme) splashModel {
	lines := strings.Split(strings.TrimRight(bannerSource, "\n"), "\n")
	total := 0
	for _, line := range lines {
		total += len(line)
	}
	return splashModel{
		theme:      theme,
		banner:     lines,
		totalChars: total,
	}
}

func (m splashModel) Init() tea.Cmd {
	return splashTick()
}

func splashTick() tea.Cmd {
	return tea.Tick(15*time.Millisecond, func(time.Time) tea.Msg { return splashTickMsg{} })
}

func (m splashModel) Update(msg tea.Msg) (splashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		m.fastForward = true
		m.done = true
		return m, func() tea.Msg { return splashDoneMsg{} }
	case splashTickMsg:
		m.revealed += 6
		if m.revealed >= m.totalChars {
			m.revealed = m.totalChars
			m.done = true
			if m.holding {
				return m, nil
			}
			m.holding = true
			return m, tea.Tick(splashHoldDuration, func(time.Time) tea.Msg { return splashDoneMsg{} })
		}
		return m, splashTick()
	}
	return m, nil
}

func (m splashModel) View() string {
	var out strings.Builder
	shown := 0
	for _, line := range m.banner {
		if shown >= m.revealed {
			out.WriteString("\n")
			continue
		}
		remaining := m.revealed - shown
		if remaining >= len(line) {
			out.WriteString(m.theme.Title.Render(line))
		} else {
			out.WriteString(m.theme.Title.Render(line[:remaining]))
		}
		out.WriteString("\n")
		shown += len(line)
	}

	subtitle := m.theme.Subtitle.Render("Klever Blockchain Localnet")
	ver := m.theme.Dim.Render(version.Full())

	content := lipgloss.JoinVertical(lipgloss.Center,
		out.String(),
		"",
		subtitle,
		ver,
	)
	if m.width == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
