package cli

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/printer"
)

type GlobalFlags struct {
	ConfigPath   string
	LogFormat    string
	LogLevel     string
	NoColor      bool
	AutoYes      bool
	ForceTUI     bool
	NoTUI        bool
	WorkDir      string
	WaitForEnter bool
}

type AppContext struct {
	Flags   *GlobalFlags
	Log     *slog.Logger
	Printer *printer.Printer
	Docker  *dockercli.Client
	WorkDir string
}

func (c *AppContext) ComposeFilePath() string {
	return filepath.Join(c.WorkDir, "docker-compose.yaml")
}

func (c *AppContext) StateFilePath() string {
	return filepath.Join(c.WorkDir, ".localnet-state.yaml")
}

func (c *AppContext) KeysDir() string    { return filepath.Join(c.WorkDir, "keys") }
func (c *AppContext) DBsDir() string     { return filepath.Join(c.WorkDir, "dbs") }
func (c *AppContext) LogsDir() string    { return filepath.Join(c.WorkDir, "logs") }
func (c *AppContext) ConfigDir() string  { return filepath.Join(c.WorkDir, "config") }
func (c *AppContext) NodeCfgDir() string { return filepath.Join(c.WorkDir, "config", "node") }

func defaultWorkDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}
