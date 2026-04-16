package cli

import (
	"github.com/spf13/cobra"
)

func newStartCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Bring the network up (docker compose up -d)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStart(cmd, app)
		},
	}
}

func runStart(cmd *cobra.Command, app *AppContext) error {
	app.Printer.Header("Starting containers")
	if _, err := app.Docker.ComposeUp(cmd.Context()); err != nil {
		return err
	}
	app.Printer.Success("Containers started. Run `localnet status` to check container status.")
	return nil
}

func newStopCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:     "stop",
		Aliases: []string{"down"},
		Short:   "Tear the network down (docker compose down)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStop(cmd, app, false)
		},
	}
}

func runStop(cmd *cobra.Command, app *AppContext, quiet bool) error {
	if !quiet {
		app.Printer.Header("Stopping containers")
	}
	if _, err := app.Docker.ComposeDown(cmd.Context()); err != nil {
		return err
	}
	if !quiet {
		app.Printer.Success("Containers stopped")
	}
	return nil
}

func newRestartCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart all services (docker compose restart)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app.Printer.Header("Restarting containers")
			if _, err := app.Docker.ComposeRestart(cmd.Context()); err != nil {
				return err
			}
			app.Printer.Success("Containers restarted")
			return nil
		},
	}
}
