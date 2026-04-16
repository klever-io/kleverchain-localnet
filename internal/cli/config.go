package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/compose"
	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
	"github.com/klever-io/kleverchain-localnet/internal/genesis"
	"github.com/klever-io/kleverchain-localnet/internal/keys"
	"github.com/klever-io/kleverchain-localnet/internal/nodes"
	"github.com/klever-io/kleverchain-localnet/internal/state"
)

type configFlags struct {
	Validators         int
	MaxSupply          uint64
	ConsensusGroupSize int
	StartTime          int64
	MinNodes           int
}

func newConfigCommand(app *AppContext) *cobra.Command {
	cf := &configFlags{}

	generate := &cobra.Command{
		Use:   "generate",
		Short: "Generate genesis.json and nodesSetup.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConfigGenerate(cmd, app, cf)
		},
	}
	bindConfigFlags(generate, cf)

	wrapper := &cobra.Command{
		Use:   "config",
		Short: "Configuration generation",
	}
	wrapper.AddCommand(generate)
	return wrapper
}

func bindConfigFlags(cmd *cobra.Command, cf *configFlags) {
	cmd.Flags().IntVarP(&cf.Validators, "validators", "n", 0, "number of validators")
	cmd.Flags().Uint64VarP(&cf.MaxSupply, "max-supply", "s", 0, "KLV max supply (default: 10^16)")
	cmd.Flags().IntVar(&cf.ConsensusGroupSize, "consensus-group-size", 0, "consensus group size (default: validators)")
	cmd.Flags().Int64Var(&cf.StartTime, "start-time", 0, "unix seconds; rounded up to next 75s boundary")
	cmd.Flags().IntVar(&cf.MinNodes, "min-nodes", 0, "minimum nodes for block production")
}

func runConfigGenerate(cmd *cobra.Command, app *AppContext, cf *configFlags) error {
	st, err := resolveStateForConfig(app, cf)
	if err != nil {
		return err
	}

	g := keys.NewGenerator(dockercli.NewExecRunner(), st.KleverImage)
	kr, err := g.Generate(cmd.Context(), st, app.KeysDir(), false)
	if err != nil {
		return err
	}

	gen, err := genesis.Build(st, kr.Wallets, kr.Root)
	if err != nil {
		return err
	}
	pairs := make([]nodes.Pair, 0, len(kr.Validators))
	for i, v := range kr.Validators {
		pairs = append(pairs, nodes.Pair{NodeIndex: v.Index, PubKey: v.PubKey, Address: kr.Wallets[i].Address})
	}
	ns, err := nodes.Build(st, pairs)
	if err != nil {
		return err
	}

	st.StartTime = ns.StartTime
	st.ChainID = ns.ChainID

	if ensureErr := fsutil.EnsureDir(app.NodeCfgDir(), 0o755); ensureErr != nil {
		return ensureErr
	}
	genesisPath := filepath.Join(app.NodeCfgDir(), "genesis.json")
	genesisBytes, marshalErr := json.MarshalIndent(gen, "", "    ")
	if marshalErr != nil {
		return marshalErr
	}
	if writeErr := fsutil.WriteFile(genesisPath, append(genesisBytes, '\n'), 0o644); writeErr != nil {
		return writeErr
	}
	if writeErr := fsutil.WriteJSON(filepath.Join(app.NodeCfgDir(), "nodesSetup.json"), ns); writeErr != nil {
		return writeErr
	}

	if saveErr := state.Save(app.StateFilePath(), st); saveErr != nil {
		return saveErr
	}
	app.Printer.Success("Wrote genesis.json and nodesSetup.json (chainID %s, startTime %d)", st.ChainID, st.StartTime)
	return nil
}

func resolveStateForConfig(app *AppContext, cf *configFlags) (domain.LocalnetState, error) {
	st, err := loadOrInitState(app, cf.Validators)
	if err != nil {
		return domain.LocalnetState{}, err
	}
	if cf.Validators > 0 {
		st.Validators = cf.Validators
	}
	if cf.MaxSupply > 0 {
		st.MaxSupply = cf.MaxSupply
	}
	if cf.ConsensusGroupSize > 0 {
		st.ConsensusGroupSize = cf.ConsensusGroupSize
	}
	if st.ConsensusGroupSize <= 0 {
		st.ConsensusGroupSize = st.Validators
	}
	if cf.MinNodes > 0 {
		st.MinNodes = cf.MinNodes
	}
	if cf.StartTime > 0 {
		st.StartTime = cf.StartTime
	}
	if st.KleverImage == "" {
		st.KleverImage = domain.KleverImage
	}
	if err := st.Validate(); err != nil {
		return domain.LocalnetState{}, fmt.Errorf("%w", err)
	}
	return st, nil
}

func runComposeGenerate(app *AppContext) error {
	if !state.Exists(app.StateFilePath()) {
		return fmt.Errorf("state file not found; run `localnet config generate` first")
	}
	st, err := state.Load(app.StateFilePath())
	if err != nil {
		return err
	}

	rel, err := fsutil.RelativeOrAbs(app.WorkDir, app.KeysDir())
	if err != nil {
		return err
	}
	services, err := compose.BuildValidatorServices(st, rel)
	if err != nil {
		return err
	}
	out, err := compose.Build(st, services, compose.DefaultResources())
	if err != nil {
		return err
	}
	if err := fsutil.WriteFile(app.ComposeFilePath(), []byte(out), 0o644); err != nil {
		return err
	}
	app.Printer.Success("Wrote %s", app.ComposeFilePath())
	return nil
}

func newComposeCommand(app *AppContext) *cobra.Command {
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Generate docker-compose.yaml",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runComposeGenerate(app)
		},
	}
	wrapper := &cobra.Command{
		Use:   "compose",
		Short: "Docker Compose file generation",
	}
	wrapper.AddCommand(generate)
	return wrapper
}
