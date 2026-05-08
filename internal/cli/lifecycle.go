package cli

import (
	"os"
	"path/filepath"
	"time"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
	"github.com/klever-io/kleverchain-localnet/internal/nodes"
	"github.com/klever-io/kleverchain-localnet/internal/state"
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

	// On a fresh-genesis run (no chain data yet), pull the validator image first
	// and then re-stamp genesis startTime. Without this, a slow image pull pushes
	// container boot past genesis and the chain stalls in inSync state forever.
	if isFreshGenesis(app) {
		if err := prepareFreshStart(cmd, app); err != nil {
			return err
		}
	}

	if _, err := app.Docker.ComposeUp(cmd.Context()); err != nil {
		return err
	}
	app.Printer.Success("Containers started. Run `localnet status` to check container status.")
	return nil
}

func isFreshGenesis(app *AppContext) bool {
	entries, err := os.ReadDir(filepath.Join(app.DBsDir(), "node-0"))
	if err != nil {
		return true
	}
	return len(entries) == 0
}

func prepareFreshStart(cmd *cobra.Command, app *AppContext) error {
	app.Printer.Info("Pulling validator image (warming cache before genesis stamp)")
	if _, err := app.Docker.ComposePull(cmd.Context()); err != nil {
		return err
	}

	nodesSetupPath := filepath.Join(app.NodeCfgDir(), "nodesSetup.json")
	var ns domain.NodesSetup
	if err := fsutil.ReadJSON(nodesSetupPath, &ns); err != nil {
		return err
	}

	newStart := nodes.RoundUpToSlotBoundary(time.Now().Unix() + domain.MinStartHeadroomSeconds)
	if newStart <= ns.StartTime {
		return nil
	}
	ns.StartTime = newStart
	if err := fsutil.WriteJSON(nodesSetupPath, ns); err != nil {
		return err
	}

	if state.Exists(app.StateFilePath()) {
		st, err := state.Load(app.StateFilePath())
		if err != nil {
			return err
		}
		st.StartTime = newStart
		if err := state.Save(app.StateFilePath(), st); err != nil {
			return err
		}
	}

	app.Printer.Info("Re-stamped genesis startTime to %d (T+%ds)", newStart, newStart-time.Now().Unix())
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
