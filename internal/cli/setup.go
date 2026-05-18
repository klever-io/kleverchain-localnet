package cli

import (
	"github.com/spf13/cobra"
)

func newSetupAllCommand(app *AppContext) *cobra.Command {
	cf := &configFlags{}
	cmd := &cobra.Command{
		Use:   "setup-all",
		Short: "One-shot: doctor + keys + dirs + config + compose",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSetupAll(cmd, app, cf)
		},
	}
	bindConfigFlags(cmd, cf)
	return cmd
}

func newRunCommand(app *AppContext) *cobra.Command {
	cf := &configFlags{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Zero-to-running: setup-all + start",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := runSetupAll(cmd, app, cf); err != nil {
				return err
			}
			return runStart(cmd, app)
		},
	}
	bindConfigFlags(cmd, cf)
	return cmd
}

func runSetupAll(cmd *cobra.Command, app *AppContext, cf *configFlags) error {
	if err := runDoctor(cmd, app, true); err != nil {
		return err
	}

	if err := runKeysGenerate(cmd, app, valueOr(cf.Validators, 1), false); err != nil {
		return err
	}
	if err := runDirsGenerate(app, cf.Validators); err != nil {
		return err
	}
	if err := runConfigGenerate(cmd, app, cf); err != nil {
		return err
	}
	if err := runComposeGenerate(app); err != nil {
		return err
	}
	printNextSteps(app)
	return nil
}

func valueOr(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

func printNextSteps(app *AppContext) {
	app.Printer.Header("Next steps")
	app.Printer.Plain("  • Start the network:    localnet start")
	app.Printer.Plain("  • Check status:         localnet status")
	app.Printer.Plain("  • Tail logs:            localnet logs --follow")
	app.Printer.Plain("  • Stop containers:      localnet stop")
}
