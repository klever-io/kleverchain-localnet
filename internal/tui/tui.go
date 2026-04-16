package tui

import (
	"context"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/printer"
)

type ActionCommand func(action string, payload any) (*exec.Cmd, error)

type Options struct {
	WorkDir       string
	Printer       *printer.Printer
	Docker        *dockercli.Client
	JumpToMonitor bool
	OnAction      ActionCommand
}

func Run(ctx context.Context, opts Options) error {
	if opts.Docker == nil {
		return fmt.Errorf("tui: docker client is required")
	}
	m := opts.buildApp(ctx)
	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
