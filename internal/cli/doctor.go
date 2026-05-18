package cli

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/printer"
)

func newDoctorCommand(app *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Verify docker, docker compose, disk space and other prerequisites",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd, app, false)
		},
	}
	return cmd
}

type doctorRow struct {
	Check  string
	Status string
	Detail string
}

func runDoctor(cmd *cobra.Command, app *AppContext, silentOnSuccess bool) error {
	app.Printer.Header("Environment check")
	rows := make([]doctorRow, 0, 6)
	fatal := false

	v, err := app.Docker.Version(cmd.Context())
	switch {
	case err != nil:
		rows = append(rows, doctorRow{"docker", "FAIL", err.Error()})
		fatal = true
	case v.Major < 20 || (v.Major == 20 && v.Minor < 10):
		rows = append(rows, doctorRow{"docker", "FAIL", fmt.Sprintf("version %d.%d is below the required 20.10", v.Major, v.Minor)})
		fatal = true
	default:
		rows = append(rows, doctorRow{"docker", "OK", v.Client})
	}

	ci, err := app.Docker.ComposeVersion(cmd.Context())
	switch {
	case err != nil:
		rows = append(rows, doctorRow{"docker compose v2", "FAIL", err.Error()})
		fatal = true
	case ci.Major < 2:
		rows = append(rows, doctorRow{"docker compose v2", "FAIL", "only compose v2 is supported (got " + ci.Raw + ")"})
		fatal = true
	default:
		rows = append(rows, doctorRow{"docker compose v2", "OK", ci.Raw})
	}

	if err := app.Docker.DaemonReachable(cmd.Context()); err != nil {
		rows = append(rows, doctorRow{"docker daemon", "FAIL", err.Error()})
		fatal = true
	} else {
		rows = append(rows, doctorRow{"docker daemon", "OK", "reachable"})
	}

	if free, err := freeDiskBytes(app.WorkDir); err != nil {
		rows = append(rows, doctorRow{"disk space", "WARN", err.Error()})
	} else {
		const minBytes uint64 = 5 * 1024 * 1024 * 1024
		status := "OK"
		if free < minBytes {
			status = "FAIL"
			fatal = true
		}
		rows = append(rows, doctorRow{"disk space", status, fmt.Sprintf("%.1f GiB free at %s", float64(free)/(1<<30), app.WorkDir)})
	}

	rows = append(rows, doctorRow{"klever image", "INFO", domain.KleverImage})

	if runtime.GOOS == "linux" {
		if err := linuxDockerGroupCheck(); err != nil {
			rows = append(rows, doctorRow{"docker group (linux)", "WARN", err.Error()})
		} else {
			rows = append(rows, doctorRow{"docker group (linux)", "OK", "member of docker group"})
		}
	}

	renderDoctorTable(app.Printer, rows)

	if fatal {
		return fmt.Errorf("%w: one or more prerequisites failed", errs.ErrPrereq)
	}
	if !silentOnSuccess {
		app.Printer.Success("All prerequisites satisfied")
	}
	return nil
}

func renderDoctorTable(p *printer.Printer, rows []doctorRow) {
	header := printer.Row{"CHECK", "STATUS", "DETAIL"}
	tableRows := make([]printer.Row, 0, len(rows))
	for _, r := range rows {
		tableRows = append(tableRows, printer.Row{r.Check, r.Status, r.Detail})
	}
	p.Table(header, tableRows)
}

func linuxDockerGroupCheck() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	gids, err := u.GroupIds()
	if err != nil {
		return err
	}
	dockerGroup, err := user.LookupGroup("docker")
	if err != nil {
		return fmt.Errorf("docker group not present on system")
	}
	for _, g := range gids {
		if g == dockerGroup.Gid {
			return nil
		}
	}
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf("user %s is not a member of the docker group", strings.TrimSpace(u.Username))
}
