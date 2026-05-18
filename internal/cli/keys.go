package cli

import (
	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
	"github.com/klever-io/kleverchain-localnet/internal/keys"
	"github.com/klever-io/kleverchain-localnet/internal/state"
)

type keysFlags struct {
	Validators int
	Force      bool
}

func newKeysCommand(app *AppContext) *cobra.Command {
	kf := &keysFlags{}
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Generate validator and wallet PEMs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runKeysGenerate(cmd, app, kf.Validators, kf.Force)
		},
	}
	generate.Flags().IntVarP(&kf.Validators, "validators", "n", 1, "number of validators")
	generate.Flags().BoolVar(&kf.Force, "force", false, "re-generate keys even if they already exist")

	wrapper := &cobra.Command{
		Use:   "keys",
		Short: "Key management",
	}
	wrapper.AddCommand(generate)
	return wrapper
}

func runKeysGenerate(cmd *cobra.Command, app *AppContext, validators int, force bool) error {
	st, err := loadOrInitState(app, validators)
	if err != nil {
		return err
	}

	if ensureErr := fsutil.EnsureDir(app.KeysDir(), 0o700); ensureErr != nil {
		return ensureErr
	}

	g := keys.NewGenerator(dockercli.NewExecRunner(), st.KleverImage)
	res, err := g.Generate(cmd.Context(), st, app.KeysDir(), force)
	if err != nil {
		return err
	}

	app.Printer.Success("Generated %d validator key set(s) and root wallet", len(res.Validators))
	for _, v := range res.Validators {
		app.Printer.Dim("  node-%d pubkey=%s", v.Index, truncate(v.PubKey, 24))
	}
	if res.Root != nil {
		app.Printer.Dim("  root wallet=%s", truncate(res.Root.Address, 24))
	}
	return state.Save(app.StateFilePath(), st)
}

func loadOrInitState(app *AppContext, validators int) (domain.LocalnetState, error) {
	if state.Exists(app.StateFilePath()) {
		loaded, err := state.Load(app.StateFilePath())
		if err != nil {
			return domain.LocalnetState{}, err
		}
		if validators > 0 && validators != loaded.Validators {
			app.Printer.Warn("existing state has %d validators; --validators=%d ignored", loaded.Validators, validators)
		}
		return loaded, nil
	}
	if validators <= 0 {
		validators = 1
	}
	return domain.NewLocalnetState(validators, domain.DefaultMaxSupply, validators), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
