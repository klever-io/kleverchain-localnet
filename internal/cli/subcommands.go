package cli

import "github.com/spf13/cobra"

func registerSubcommands(root *cobra.Command, app *AppContext) {
	root.AddCommand(
		newRunCommand(app),
		newSetupAllCommand(app),
		newDoctorCommand(app),
		newKeysCommand(app),
		newDirsCommand(app),
		newConfigCommand(app),
		newComposeCommand(app),
		newStartCommand(app),
		newStopCommand(app),
		newRestartCommand(app),
		newStatusCommand(app),
		newLogsCommand(app),
		newMonitorCommand(app),
		newCleanCommand(app),
		newCleanAllCommand(app),
		newVersionCommand(app),
	)
}
