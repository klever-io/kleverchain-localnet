package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/klever-io/kleverchain-localnet/internal/detect"
	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
)

type screen int

const (
	screenSplash screen = iota
	screenMenu
	screenMonitor
	screenLogs
	screenSetup
	screenConfirm
	screenHelp
)

type app struct {
	ctx           context.Context
	theme         *Theme
	keymap        KeyMap
	docker        *dockercli.Client
	workDir       string
	screen        screen
	splash        splashModel
	menu          menuModel
	monitor       monitorModel
	logs          logsModel
	setupForm     setupFormModel
	confirm       confirmModel
	help          helpModel
	prevScreen    screen
	pendingAction menuAction
	detector      detect.Detector
	width         int
	height        int
	lastSnap      detect.Snapshot
	lastErr       error
	initComplete  bool
	jumpMonitor   bool
	onAction      ActionCommand
}

type detectedMsg struct{ snap detect.Snapshot }
type actionDoneMsg struct {
	action menuAction
	err    error
}

func New(ctx context.Context, docker *dockercli.Client, workDir string, jumpToMonitor bool, onAction ActionCommand) *app {
	theme := NewTheme()
	keymap := DefaultKeyMap()
	a := &app{
		ctx:         ctx,
		theme:       theme,
		keymap:      keymap,
		docker:      docker,
		workDir:     workDir,
		screen:      screenSplash,
		splash:      newSplash(theme),
		menu:        newMenu(theme, keymap, detect.Snapshot{}),
		detector:    detect.Detector{WorkDir: workDir, Docker: &dockerProbe{client: docker}},
		jumpMonitor: jumpToMonitor,
		onAction:    onAction,
	}
	if jumpToMonitor {
		a.screen = screenMonitor
		a.monitor = newMonitor(theme, docker, 2*time.Second, false)
	}
	return a
}

type dockerProbe struct {
	client *dockercli.Client
}

func (p *dockerProbe) ListServices(ctx context.Context) ([]detect.DockerService, error) {
	services, err := p.client.ComposePS(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]detect.DockerService, 0, len(services))
	for _, s := range services {
		out = append(out, detect.DockerService{Name: s.Name, State: s.State, Health: s.Health})
	}
	return out, nil
}

func (a *app) Init() tea.Cmd {
	cmds := []tea.Cmd{a.detectCmd()}
	if a.screen == screenSplash {
		cmds = append(cmds, a.splash.Init())
	} else if a.screen == screenMonitor {
		cmds = append(cmds, a.monitor.Init())
	}
	return tea.Batch(cmds...)
}

func (a *app) detectCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		snap := a.detector.Detect(ctx)
		return detectedMsg{snap: snap}
	}
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if a.screen != screenHelp && a.screen != screenSetup && a.screen != screenLogs && m.String() == "h" {
			a.prevScreen = a.screen
			a.screen = screenHelp
			a.help = newHelp(a.theme)
			return a, nil
		}
	case detectedMsg:
		a.lastSnap = m.snap
		a.menu = a.menu.updateSnapshot(m.snap)
		a.initComplete = true
	case splashDoneMsg:
		a.screen = screenMenu
		return a, nil
	case menuSelectMsg:
		return a.handleMenuAction(m.action)
	case setupFormSubmitMsg:
		a.screen = screenMenu
		return a, a.runActionAsync(actionSetup, m.values)
	case setupFormCancelMsg:
		a.screen = screenMenu
	case actionDoneMsg:
		if m.err != nil {
			a.lastErr = m.err
		}
		return a, a.detectCmd()
	case confirmAnswerMsg:
		a.screen = screenMenu
		if m.Result {
			switch a.pendingAction {
			case actionStop:
				return a, a.runActionAsync(actionStop, nil)
			case actionRestart:
				return a, a.runActionAsync(actionRestart, nil)
			case actionClean:
				return a, a.runActionAsync(actionClean, nil)
			case actionCleanAll:
				return a, a.runActionAsync(actionCleanAll, nil)
			}
		}
	case monitorExitMsg:
		a.screen = screenMenu
		return a, a.detectCmd()
	case logsExitMsg:
		a.screen = screenMenu
		return a, a.detectCmd()
	case helpExitMsg:
		if a.prevScreen == 0 {
			a.screen = screenMenu
		} else {
			a.screen = a.prevScreen
		}
		return a, nil
	}

	var cmd tea.Cmd
	switch a.screen {
	case screenSplash:
		a.splash, cmd = a.splash.Update(msg)
	case screenMenu:
		a.menu, cmd = a.menu.Update(msg)
	case screenMonitor:
		a.monitor, cmd = a.monitor.Update(msg)
	case screenLogs:
		a.logs, cmd = a.logs.Update(msg)
	case screenSetup:
		a.setupForm, cmd = a.setupForm.Update(msg)
	case screenConfirm:
		a.confirm, cmd = a.confirm.Update(msg)
	case screenHelp:
		a.help, cmd = a.help.Update(msg)
	}
	return a, cmd
}

func (a *app) handleMenuAction(action menuAction) (tea.Model, tea.Cmd) {
	switch action {
	case actionQuit:
		return a, tea.Quit
	case actionSetup:
		a.screen = screenSetup
		a.setupForm = newSetupForm(a.theme)
		return a, a.setupForm.Init()
	case actionStart:
		return a, a.runActionAsync(actionStart, nil)
	case actionStop:
		a.pendingAction = actionStop
		a.screen = screenConfirm
		a.confirm = newConfirm(a.theme, "stop", "Stop network?", "This will bring all containers down.")
		return a, nil
	case actionRestart:
		a.pendingAction = actionRestart
		a.screen = screenConfirm
		a.confirm = newConfirm(a.theme, "restart", "Restart all services?", "docker compose restart")
		return a, nil
	case actionMonitor:
		a.screen = screenMonitor
		a.monitor = newMonitor(a.theme, a.docker, 2*time.Second, false)
		return a, a.monitor.Init()
	case actionLogs:
		a.screen = screenLogs
		a.logs = newLogs(a.theme, a.docker, "")
		return a, a.logs.Init()
	case actionClean:
		a.pendingAction = actionClean
		a.screen = screenConfirm
		a.confirm = newConfirm(a.theme, "clean", "Remove configs?", "Removes genesis, nodesSetup, and compose files.")
		return a, nil
	case actionCleanAll:
		a.pendingAction = actionCleanAll
		a.screen = screenConfirm
		a.confirm = newConfirm(a.theme, "clean-all", "DESTRUCTIVE: nuke everything?", "Removes keys, dbs, logs, and configs.")
		return a, nil
	case actionDoctor:
		return a, a.runActionAsync(actionDoctor, nil)
	}
	return a, nil
}

func (a *app) runActionAsync(action menuAction, payload any) tea.Cmd {
	handler := a.onAction
	if handler == nil {
		return func() tea.Msg {
			return actionDoneMsg{action: action, err: fmt.Errorf("no action handler configured")}
		}
	}
	execCmd, err := handler(string(action), payload)
	if err != nil {
		return func() tea.Msg {
			return actionDoneMsg{action: action, err: err}
		}
	}
	if execCmd == nil {
		return func() tea.Msg {
			return actionDoneMsg{action: action, err: fmt.Errorf("action %q produced no command", action)}
		}
	}
	if execCmd.Stdout == nil {
		execCmd.Stdout = os.Stdout
	}
	if execCmd.Stderr == nil {
		execCmd.Stderr = os.Stderr
	}
	if execCmd.Stdin == nil {
		execCmd.Stdin = os.Stdin
	}
	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return actionDoneMsg{action: action, err: err}
	})
}

func (a *app) View() string {
	switch a.screen {
	case screenSplash:
		return a.splash.View()
	case screenMenu:
		return a.menu.View()
	case screenMonitor:
		return a.monitor.View()
	case screenLogs:
		return a.logs.View()
	case screenSetup:
		return a.setupForm.View()
	case screenConfirm:
		return a.confirm.View()
	case screenHelp:
		return a.help.View()
	}
	return ""
}

func (o Options) buildApp(ctx context.Context) *app {
	return New(ctx, o.Docker, o.WorkDir, o.JumpToMonitor, o.OnAction)
}
