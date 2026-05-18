package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/printer"
	"github.com/klever-io/kleverchain-localnet/internal/version"
)

func Execute() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	root, app := buildRootCommand()
	root.SetContext(ctx)

	err := root.Execute()
	code := 0
	if err != nil {
		if app != nil && app.Printer != nil {
			app.Printer.Error("%s", err.Error())
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		code = exitCodeFor(err)
	}

	if app != nil && app.Flags != nil && app.Flags.WaitForEnter {
		waitForEnter()
	}
	return code
}

func waitForEnter() {
	fmt.Fprint(os.Stdout, "\nPress Enter to return to the menu... ")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func buildRootCommand() (*cobra.Command, *AppContext) {
	gf := &GlobalFlags{}
	app := &AppContext{
		Flags:   gf,
		WorkDir: defaultWorkDir(),
	}

	root := &cobra.Command{
		Use:           "localnet",
		Short:         "Klever Blockchain Localnet — run a multi-validator testnet locally",
		Long:          "localnet is a single-binary tool for bootstrapping, running, and monitoring a local Klever blockchain network.",
		Version:       version.Full(),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initApp(app)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return dispatchBareInvocation(cmd, app)
		},
	}

	root.PersistentFlags().StringVar(&gf.ConfigPath, "config", "", "optional YAML config file")
	root.PersistentFlags().StringVar(&gf.LogFormat, "log-format", "text", "log format: text|json")
	root.PersistentFlags().StringVar(&gf.LogLevel, "log-level", "info", "log level: debug|info|warn|error")
	root.PersistentFlags().BoolVar(&gf.NoColor, "no-color", false, "disable ANSI colors")
	root.PersistentFlags().BoolVarP(&gf.AutoYes, "yes", "y", false, "auto-confirm destructive operations")
	root.PersistentFlags().BoolVar(&gf.ForceTUI, "tui", false, "force-launch the TUI (bare invocation only)")
	root.PersistentFlags().BoolVar(&gf.NoTUI, "no-tui", false, "force CLI help even on a TTY (bare invocation only)")
	root.PersistentFlags().StringVar(&gf.WorkDir, "work-dir", "", "working directory (default: cwd)")
	root.PersistentFlags().BoolVar(&gf.WaitForEnter, "wait-for-enter", false, "after the command completes, pause until Enter is pressed (used internally by the TUI)")
	_ = root.PersistentFlags().MarkHidden("wait-for-enter")

	root.SetVersionTemplate("{{.Version}}\n")

	registerSubcommands(root, app)
	return root, app
}

func initApp(app *AppContext) error {
	if app.Flags.WorkDir != "" {
		app.WorkDir = app.Flags.WorkDir
	}
	app.Printer = printer.New(os.Stdout, os.Stderr, app.Flags.NoColor)

	level := parseLogLevel(app.Flags.LogLevel)
	var handler slog.Handler
	switch strings.ToLower(app.Flags.LogFormat) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	app.Log = slog.New(handler)
	slog.SetDefault(app.Log)

	app.Docker = dockercli.New(
		dockercli.WithWorkDir(app.WorkDir),
	)
	return nil
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func dispatchBareInvocation(cmd *cobra.Command, app *AppContext) error {
	if app.Flags.NoTUI {
		return cmd.Help()
	}
	if app.Flags.ForceTUI {
		if !stdinIsTTY() {
			return fmt.Errorf("%w: --tui requires an interactive terminal", errs.ErrPrereq)
		}
		return runTUI(cmd.Context(), app, false)
	}
	if stdoutIsTTY() && stdinIsTTY() && os.Getenv("TERM") != "dumb" {
		return runTUI(cmd.Context(), app, false)
	}
	return cmd.Help()
}

func stdoutIsTTY() bool { return isatty.IsTerminal(os.Stdout.Fd()) }
func stdinIsTTY() bool  { return isatty.IsTerminal(os.Stdin.Fd()) }

func exitCodeFor(err error) int {
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		return 2
	case errors.Is(err, errs.ErrPrereq):
		return 3
	case errors.Is(err, errs.ErrStateConflict), errors.Is(err, errs.ErrKeysIncomplete):
		return 4
	case errors.Is(err, errs.ErrDockerFailure), errors.Is(err, errs.ErrComposeMissing):
		return 5
	}
	return 1
}
