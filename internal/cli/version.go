package cli

import (
	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/version"
)

func newVersionCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit SHA, and build date",
		RunE: func(_ *cobra.Command, _ []string) error {
			app.Printer.Plain("localnet %s", version.Full())
			return nil
		},
	}
}
