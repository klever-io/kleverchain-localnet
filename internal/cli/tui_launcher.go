package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/klever-io/kleverchain-localnet/internal/tui"
)

type setupFormValuesLike interface {
	GetValidators() int
	GetMaxSupply() uint64
	GetConsensusGroupSize() int
}

func runTUI(ctx context.Context, app *AppContext, jumpToMonitor bool) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate self executable: %w", err)
	}

	builder := func(action string, payload any) (*exec.Cmd, error) {
		args, err := resolveTUIAction(action, payload)
		if err != nil {
			return nil, err
		}
		globalArgs := []string{"--work-dir", app.WorkDir, "--yes", "--wait-for-enter"}
		if app.Flags.NoColor {
			globalArgs = append(globalArgs, "--no-color")
		}
		if app.Flags.LogFormat != "" && app.Flags.LogFormat != "text" {
			globalArgs = append(globalArgs, "--log-format", app.Flags.LogFormat)
		}
		full := make([]string, 0, len(globalArgs)+len(args))
		full = append(full, globalArgs...)
		full = append(full, args...)
		cmd := exec.CommandContext(ctx, self, full...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd, nil
	}

	opts := tui.Options{
		WorkDir:       app.WorkDir,
		Printer:       app.Printer,
		Docker:        app.Docker,
		JumpToMonitor: jumpToMonitor,
		OnAction:      builder,
	}
	return tui.Run(ctx, opts)
}

func resolveTUIAction(action string, payload any) ([]string, error) {
	switch action {
	case "setup":
		args := []string{"setup-all"}
		if vals, ok := payload.(setupFormValuesLike); ok {
			args = append(args,
				"-n", strconv.Itoa(vals.GetValidators()),
				"-s", strconv.FormatUint(vals.GetMaxSupply(), 10),
				"--consensus-group-size", strconv.Itoa(vals.GetConsensusGroupSize()),
			)
		}
		return args, nil
	case "start":
		return []string{"start"}, nil
	case "stop":
		return []string{"stop"}, nil
	case "restart":
		return []string{"restart"}, nil
	case "clean":
		return []string{"clean"}, nil
	case "clean-all":
		return []string{"clean-all"}, nil
	case "doctor":
		return []string{"doctor"}, nil
	}
	return nil, fmt.Errorf("unknown TUI action %q", action)
}
