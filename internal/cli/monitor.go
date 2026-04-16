package cli

import "github.com/spf13/cobra"

func newMonitorCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Open the TUI monitor dashboard",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTUI(cmd.Context(), app, true)
		},
	}
}
