package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

func newCleanCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove generated configs and compose file (keys/dbs/logs preserved)",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runClean(app)
		},
	}
}

func runClean(app *AppContext) error {
	paths := []string{
		filepath.Join(app.WorkDir, "docker-compose.yaml"),
		filepath.Join(app.WorkDir, "docker-compose.yml"),
		filepath.Join(app.NodeCfgDir(), "genesis.json"),
		filepath.Join(app.NodeCfgDir(), "nodesSetup.json"),
		filepath.Join(app.WorkDir, "configs"),
		app.StateFilePath(),
	}
	for _, p := range paths {
		if err := fsutil.RemoveIfExists(p); err != nil {
			return err
		}
	}
	app.Printer.Success("Removed generated configs")
	return nil
}

func newCleanAllCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "clean-all",
		Short: "Remove keys, dbs, logs, configs, and compose file (destructive)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCleanAll(cmd, app)
		},
	}
}

func runCleanAll(cmd *cobra.Command, app *AppContext) error {
	if !app.Flags.AutoYes {
		app.Printer.Warn("This will remove keys, databases, logs, configs, and the compose file.")
		confirmed, err := promptYesNo("Are you sure you want to continue? (yes/no): ")
		if err != nil {
			return err
		}
		if !confirmed {
			app.Printer.Info("Aborted.")
			return nil
		}
	}

	if err := runStop(cmd, app, true); err != nil {
		app.Printer.Warn("stop returned an error (continuing): %v", err)
	}

	if err := runClean(app); err != nil {
		return err
	}
	paths := []string{
		app.KeysDir(),
		app.DBsDir(),
		app.LogsDir(),
	}
	for _, p := range paths {
		if err := fsutil.RemoveIfExists(p); err != nil {
			return err
		}
	}
	app.Printer.Success("Removed keys, databases, logs, and generated configs")
	return nil
}

func promptYesNo(prompt string) (bool, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return ans == "yes" || ans == "y", nil
}
