package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
)

type setupFormValues struct {
	Validators         int
	MaxSupply          uint64
	ConsensusGroupSize int
}

func (v setupFormValues) GetValidators() int         { return v.Validators }
func (v setupFormValues) GetMaxSupply() uint64       { return v.MaxSupply }
func (v setupFormValues) GetConsensusGroupSize() int { return v.ConsensusGroupSize }

type setupFormModel struct {
	theme   *Theme
	inputs  []textinput.Model
	focus   int
	err     string
	result  *setupFormValues
	width   int
	height  int
	aborted bool
}

type setupFormSubmitMsg struct {
	values setupFormValues
}
type setupFormCancelMsg struct{}

func newSetupForm(theme *Theme) setupFormModel {
	validators := textinput.New()
	validators.Placeholder = "3"
	validators.Prompt = "Validators (1–10): "
	validators.SetValue("3")
	validators.CharLimit = 3

	supply := textinput.New()
	supply.Placeholder = fmt.Sprintf("%d", domain.DefaultMaxSupply)
	supply.Prompt = "Max supply: "
	supply.SetValue(fmt.Sprintf("%d", domain.DefaultMaxSupply))
	supply.CharLimit = 20

	consensus := textinput.New()
	consensus.Placeholder = "3"
	consensus.Prompt = "Consensus group size (≤ validators): "
	consensus.SetValue("3")
	consensus.CharLimit = 3

	validators.Focus()

	return setupFormModel{
		theme:  theme,
		inputs: []textinput.Model{validators, supply, consensus},
	}
}

func (m setupFormModel) Init() tea.Cmd { return textinput.Blink }

func (m setupFormModel) Update(msg tea.Msg) (setupFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.aborted = true
			return m, func() tea.Msg { return setupFormCancelMsg{} }
		case "tab", "down":
			m.focusNext()
		case "shift+tab", "up":
			m.focusPrev()
		case "enter":
			if m.focus == len(m.inputs)-1 {
				return m.submit()
			}
			m.focusNext()
		}
	}
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		var c tea.Cmd
		m.inputs[i], c = m.inputs[i].Update(msg)
		cmds[i] = c
	}
	return m, tea.Batch(cmds...)
}

func (m *setupFormModel) focusNext() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
}

func (m *setupFormModel) focusPrev() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
}

func (m setupFormModel) submit() (setupFormModel, tea.Cmd) {
	v, err := strconv.Atoi(strings.TrimSpace(m.inputs[0].Value()))
	if err != nil || v < 1 || v > 10 {
		m.err = "Validators must be an integer between 1 and 10"
		return m, nil
	}
	s, err := strconv.ParseUint(strings.TrimSpace(m.inputs[1].Value()), 10, 64)
	if err != nil || s == 0 {
		m.err = "Max supply must be a positive integer"
		return m, nil
	}
	c, err := strconv.Atoi(strings.TrimSpace(m.inputs[2].Value()))
	if err != nil || c < 1 || c > v {
		m.err = fmt.Sprintf("Consensus group size must be in [1, %d]", v)
		return m, nil
	}
	totalStaking := domain.KLVDelegation * uint64(v)
	if s <= totalStaking+domain.RootKLV {
		m.err = "Max supply must exceed totalStaking + rootKLV"
		return m, nil
	}
	m.err = ""
	m.result = &setupFormValues{Validators: v, MaxSupply: s, ConsensusGroupSize: c}
	return m, func() tea.Msg { return setupFormSubmitMsg{values: *m.result} }
}

func (m setupFormModel) View() string {
	title := m.theme.Title.Render("Setup-all configuration")
	hint := m.theme.Help.Render("enter submit · esc cancel · tab/shift+tab navigate")
	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	for _, in := range m.inputs {
		lines = append(lines, in.View())
	}
	if m.err != "" {
		lines = append(lines, "")
		lines = append(lines, m.theme.ErrStyle.Render(m.err))
	}
	lines = append(lines, "")
	lines = append(lines, hint)
	panel := m.theme.Panel.Width(72).Render(strings.Join(lines, "\n"))
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
