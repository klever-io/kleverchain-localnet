package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/klever-io/kleverchain-localnet/internal/detect"
	"github.com/klever-io/kleverchain-localnet/internal/version"
)

type menuAction string

const (
	actionSetup    menuAction = "setup"
	actionStart    menuAction = "start"
	actionStop     menuAction = "stop"
	actionRestart  menuAction = "restart"
	actionMonitor  menuAction = "monitor"
	actionLogs     menuAction = "logs"
	actionClean    menuAction = "clean"
	actionCleanAll menuAction = "clean-all"
	actionDoctor   menuAction = "doctor"
	actionQuit     menuAction = "quit"
)

type menuItem struct {
	hotkey     string
	label      string
	desc       string
	action     menuAction
	enabledFor map[detect.State]bool
}

func (m menuItem) enabled(state detect.State) bool {
	if m.enabledFor == nil {
		return true
	}
	return m.enabledFor[state]
}

type menuModel struct {
	theme     *Theme
	keymap    KeyMap
	items     []menuItem
	cursor    int
	snap      detect.Snapshot
	width     int
	height    int
	statusMsg string
}

func newMenu(theme *Theme, keymap KeyMap, snap detect.Snapshot) menuModel {
	items := defaultMenuItems()
	return menuModel{
		theme:  theme,
		keymap: keymap,
		items:  items,
		snap:   snap,
	}
}

func defaultMenuItems() []menuItem {
	setup := map[detect.State]bool{detect.Uninitialized: true, detect.KeysOnly: true, detect.Partial: true, detect.Ready: true}
	start := map[detect.State]bool{detect.Ready: true}
	running := map[detect.State]bool{detect.Running: true}
	noRun := map[detect.State]bool{
		detect.Uninitialized: true, detect.KeysOnly: true, detect.Partial: true, detect.Ready: true,
	}
	return []menuItem{
		{"s", "Setup-all", "Generate keys, configs, and compose", actionSetup, setup},
		{"u", "Start", "Bring the network up", actionStart, start},
		{"d", "Stop", "Tear the network down", actionStop, running},
		{"r", "Restart", "docker compose restart", actionRestart, running},
		{"m", "Monitor", "Live node dashboard", actionMonitor, running},
		{"l", "Logs", "Tail logs from a node", actionLogs, running},
		{"x", "Clean", "Remove generated configs", actionClean, noRun},
		{"X", "Clean-all", "Nuke keys/dbs/logs (confirm)", actionCleanAll, nil},
		{"?", "Doctor", "Re-run requirement checks", actionDoctor, nil},
		{"q", "Quit", "", actionQuit, nil},
	}
}

type menuSelectMsg struct {
	action menuAction
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case msg.String() == "up", msg.String() == "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case msg.String() == "down", msg.String() == "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case msg.String() == "enter":
			item := m.items[m.cursor]
			if !item.enabled(m.snap.State) {
				m.statusMsg = fmt.Sprintf("[%s] not available in state %s", item.hotkey, m.snap.State)
				return m, nil
			}
			return m, func() tea.Msg { return menuSelectMsg{action: item.action} }
		default:
			for _, item := range m.items {
				if item.hotkey == msg.String() {
					if !item.enabled(m.snap.State) {
						m.statusMsg = fmt.Sprintf("[%s] not available in state %s", item.hotkey, m.snap.State)
						return m, nil
					}
					return m, func() tea.Msg { return menuSelectMsg{action: item.action} }
				}
			}
		}
	}
	return m, nil
}

func (m menuModel) updateSnapshot(snap detect.Snapshot) menuModel {
	m.snap = snap
	return m
}

func (m menuModel) View() string {
	title := m.theme.Title.Render("KLOCAL — Localnet Control")
	versionStr := m.theme.Dim.Render("v" + trimDev(version.Version))
	badge := renderBadge(m.theme, m.snap)
	nextStep := renderNextStepBanner(m.theme, m.snap)

	var lines []string
	lines = append(lines, title+"  "+versionStr)
	lines = append(lines, "")
	lines = append(lines, badge)
	lines = append(lines, "")
	lines = append(lines, nextStep)
	lines = append(lines, "")

	for i, item := range m.items {
		prefix := "  "
		if i == m.cursor {
			prefix = m.theme.Selected.Render("▶ ")
		}

		var body string
		if item.enabled(m.snap.State) {
			hotkey := m.theme.Hotkey.Render("[" + item.hotkey + "]")
			body = fmt.Sprintf("%s %-14s %s", hotkey, item.label, m.theme.Dim.Render(item.desc))
		} else {
			body = m.theme.Dim.Render(fmt.Sprintf("[%s] %-14s %s", item.hotkey, item.label, item.desc))
		}

		lines = append(lines, prefix+body)
	}

	if m.statusMsg != "" {
		lines = append(lines, "")
		lines = append(lines, m.theme.Warn.Render(m.statusMsg))
	}

	help := m.theme.Help.Render("↑/↓ navigate · enter select · h help · q quit")
	lines = append(lines, "")
	lines = append(lines, help)

	panel := m.theme.Panel.Width(78).Render(strings.Join(lines, "\n"))
	if m.width == 0 {
		return panel
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func trimDev(v string) string {
	if strings.HasPrefix(v, "v") {
		return v[1:]
	}
	return v
}

func renderBadge(th *Theme, snap detect.Snapshot) string {
	switch snap.State {
	case detect.Uninitialized:
		return th.Dim.Render("○ Not initialized")
	case detect.KeysOnly:
		return th.Warn.Render("◐ Keys generated")
	case detect.Ready:
		return th.Success.Render("◑ Ready to start")
	case detect.Running:
		txt := fmt.Sprintf("● Running   %d/%d nodes healthy", snap.HealthyServices, snap.RunningServices)
		return th.BadgeOK.Render(txt)
	case detect.Partial:
		return th.BadgeWarn.Render("⚠ Partial state — run setup-all")
	}
	return th.Dim.Render("unknown state")
}

func renderNextStepBanner(th *Theme, snap detect.Snapshot) string {
	var text string
	switch snap.State {
	case detect.Uninitialized:
		text = "No keys found. Press [s] to run setup-all."
	case detect.KeysOnly:
		text = "Keys exist but no compose. Press [s] to finish setup."
	case detect.Ready:
		text = "Everything generated. Press [u] to start the network."
	case detect.Running:
		text = "Press [m] to open the live monitor."
	case detect.Partial:
		text = "Detected inconsistent state. Press [X] then [s] to reset."
	}
	return th.Subtitle.Render("➜ Suggested next: ") + text
}
