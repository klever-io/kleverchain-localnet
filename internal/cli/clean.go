package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

// chownToHostOnLinux is a Linux-only pre-cleanup step. Validator containers run as
// image-default `klever` (uid 999) and write into the bind-mounted dbs/ and logs/
// trees, leaving files owned by uid 999 on the host. The host wrapper user (typically
// uid 1000) then cannot recurse into 0700 subdirectories or unlink files inside them,
// making `os.RemoveAll` fail with `permission denied`. Resolve by running a throwaway
// container as root that chowns the target paths back to the host uid/gid before the
// host-side RemoveAll runs. No-op on Windows/macOS — Docker Desktop's filesystem
// shim does not preserve in-container uid/gid on bind-mounts there.
func chownToHostOnLinux(ctx context.Context, paths []string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	uid := strconv.Itoa(os.Getuid())
	gid := strconv.Itoa(os.Getgid())
	for _, p := range paths {
		if !fsutil.Exists(p) {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve %s: %w", p, err)
		}
		args := []string{
			"run", "--rm",
			"--user", "0:0",
			"-v", abs + ":/work",
			"--entrypoint", "chown",
			domain.KleverImage,
			"-R", uid + ":" + gid, "/work",
		}
		cmd := exec.CommandContext(ctx, "docker", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("chown %s via docker: %w (%s)", p, err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

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
	if err := chownToHostOnLinux(cmd.Context(), []string{app.DBsDir(), app.LogsDir()}); err != nil {
		app.Printer.Warn("chown-to-host pre-cleanup returned an error (continuing): %v", err)
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
