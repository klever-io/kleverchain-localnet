package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/printer"
)

func newStatusCommand(app *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show container status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := app.Docker.ComposePS(cmd.Context())
			if err != nil {
				return err
			}
			if len(services) == 0 {
				app.Printer.Info("No containers found. Run `localnet start`.")
				return nil
			}
			header := printer.Row{"SERVICE", "CONTAINER", "STATE", "HEALTH", "PORTS"}
			rows := make([]printer.Row, 0, len(services))
			for _, s := range services {
				rows = append(rows, printer.Row{
					s.ServiceName,
					s.Name,
					s.State,
					valOrDash(s.Health),
					formatPorts(s),
				})
			}
			app.Printer.Header("Container status")
			app.Printer.Table(header, rows)
			return nil
		},
	}
}

func valOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func formatPorts(s dockercli.Service) string {
	if len(s.Ports) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(s.Ports))
	seen := make(map[string]struct{})
	for _, p := range s.Ports {
		if p.Host == "" {
			continue
		}
		if _, ok := seen[p.Host]; ok {
			continue
		}
		seen[p.Host] = struct{}{}
		parts = append(parts, p.Host)
	}
	return strings.Join(parts, ",")
}
