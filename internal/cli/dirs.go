package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

func newDirsCommand(app *AppContext) *cobra.Command {
	var validators int

	generate := &cobra.Command{
		Use:   "generate",
		Short: "Create dbs/ and logs/ per-validator directories",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDirsGenerate(app, validators)
		},
	}
	generate.Flags().IntVarP(&validators, "validators", "n", 0, "number of validators (defaults to state)")

	wrapper := &cobra.Command{
		Use:   "dirs",
		Short: "Directory management",
	}
	wrapper.AddCommand(generate)
	return wrapper
}

func runDirsGenerate(app *AppContext, validators int) error {
	st, err := loadOrInitState(app, validators)
	if err != nil {
		return err
	}

	for i := 0; i < st.Validators; i++ {
		db := filepath.Join(app.DBsDir(), fmt.Sprintf("node-%d", i))
		lg := filepath.Join(app.LogsDir(), fmt.Sprintf("node-%d", i))
		if err := fsutil.EnsureDir(db, 0o755); err != nil {
			return err
		}
		if err := fsutil.EnsureDir(lg, 0o755); err != nil {
			return err
		}
	}
	app.Printer.Success("Created directories for %d node(s)", st.Validators)
	return nil
}
