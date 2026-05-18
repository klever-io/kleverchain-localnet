package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
)

type logsFlags struct {
	Follow   bool
	NoFollow bool
	Node     int
	Tail     int
}

func newLogsCommand(app *AppContext) *cobra.Command {
	lf := &logsFlags{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Tail container logs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			follow := lf.Follow
			if lf.NoFollow {
				follow = false
			}
			node := ""
			if lf.Node >= 0 {
				node = fmt.Sprintf("node%d", lf.Node)
			}
			return app.Docker.ComposeLogs(cmd.Context(), dockercli.ComposeLogOpts{
				Follow: follow,
				Node:   node,
				Tail:   lf.Tail,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			})
		},
	}
	cmd.Flags().BoolVarP(&lf.Follow, "follow", "f", true, "follow log output")
	cmd.Flags().BoolVar(&lf.NoFollow, "no-follow", false, "disable following (one-shot output)")
	cmd.Flags().IntVarP(&lf.Node, "node", "N", -1, "limit logs to a single node index (e.g. 0 → node0)")
	cmd.Flags().IntVar(&lf.Tail, "tail", 0, "last N lines (0 = all)")
	return cmd
}
