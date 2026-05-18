package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	theme  *Theme
	title  string
	body   string
	width  int
	height int
	id     string
}

type confirmAnswerMsg struct {
	ID     string
	Result bool
}

func newConfirm(theme *Theme, id, title, body string) confirmModel {
	return confirmModel{theme: theme, title: title, body: body, id: id}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "y":
			return m, func() tea.Msg { return confirmAnswerMsg{ID: m.id, Result: true} }
		case "n", "esc", "q":
			return m, func() tea.Msg { return confirmAnswerMsg{ID: m.id, Result: false} }
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	var lines []string
	lines = append(lines, m.theme.Title.Render(m.title))
	lines = append(lines, "")
	lines = append(lines, m.body)
	lines = append(lines, "")
	lines = append(lines, m.theme.Help.Render("[y] yes · [n] no"))
	panel := m.theme.Panel.Width(60).Render(strings.Join(lines, "\n"))
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
