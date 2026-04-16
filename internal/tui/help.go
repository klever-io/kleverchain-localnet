package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	theme  *Theme
	width  int
	height int
}

type helpExitMsg struct{}

func newHelp(theme *Theme) helpModel {
	return helpModel{theme: theme}
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "h", "enter":
			return m, func() tea.Msg { return helpExitMsg{} }
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"↑/k", "move up"},
				{"↓/j", "move down"},
				{"enter", "select"},
				{"esc", "back"},
				{"q", "quit / back"},
				{"ctrl+c", "force quit"},
			},
		},
		{
			title: "Main menu actions",
			keys: [][2]string{
				{"s", "Setup-all (keys + configs + compose)"},
				{"u", "Start the network"},
				{"d", "Stop the network"},
				{"r", "Restart all services"},
				{"m", "Monitor live containers"},
				{"l", "Tail node logs"},
				{"x", "Clean (remove generated configs)"},
				{"X", "Clean-all (wipe keys/dbs/logs)"},
				{"?", "Doctor (re-run requirement checks)"},
				{"h", "Toggle this help overlay"},
			},
		},
		{
			title: "Monitor screen",
			keys: [][2]string{
				{"↑/↓", "select row"},
				{"q/esc", "back to menu"},
			},
		},
		{
			title: "Logs screen",
			keys: [][2]string{
				{"g", "jump to top"},
				{"G", "jump to bottom"},
				{"f", "toggle follow"},
				{"q/esc", "back to menu"},
			},
		},
	}

	var b strings.Builder
	b.WriteString(m.theme.Title.Render("KLOCAL — Help"))
	b.WriteString("\n\n")
	for i, sec := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(m.theme.Subtitle.Render(sec.title))
		b.WriteString("\n")
		for _, kv := range sec.keys {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				m.theme.Hotkey.Render(fmt.Sprintf("%-8s", kv[0])),
				m.theme.Dim.Render(kv[1]),
			))
		}
	}
	b.WriteString("\n")
	b.WriteString(m.theme.Help.Render("press esc, h, or enter to close"))

	panel := m.theme.Panel.Width(72).Render(b.String())
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
