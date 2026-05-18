package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

// nodeDataDirMode is the permission applied to per-validator `dbs/` and `logs/` directories.
// On Linux it must be world-writable: the validator container runs as image-default `klever`
// (uid 999), but the host wrapper runs as the invoking user (typically uid 1000), so the
// in-container process needs explicit "other" write access. On Windows/macOS Docker Desktop
// the host filesystem layer does not enforce these POSIX bits on bind-mounts, so 0755 is fine.
func nodeDataDirMode() fs.FileMode {
	if runtime.GOOS == "linux" {
		return 0o777
	}
	return 0o755
}

// ensureNodeDataDir creates the directory and explicitly chmods it, bypassing umask.
// fsutil.EnsureDir uses MkdirAll which masks the requested mode against the process umask
// (typically 022 → strips group/other write), defeating the 0777 we need on Linux.
func ensureNodeDataDir(path string, mode fs.FileMode) error {
	if err := fsutil.EnsureDir(path, mode); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

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

	mode := nodeDataDirMode()
	for i := 0; i < st.Validators; i++ {
		db := filepath.Join(app.DBsDir(), fmt.Sprintf("node-%d", i))
		lg := filepath.Join(app.LogsDir(), fmt.Sprintf("node-%d", i))
		if err := ensureNodeDataDir(db, mode); err != nil {
			return err
		}
		if err := ensureNodeDataDir(lg, mode); err != nil {
			return err
		}
	}
	app.Printer.Success("Created directories for %d node(s)", st.Validators)
	return nil
}
