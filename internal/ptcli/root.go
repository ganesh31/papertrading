package ptcli

import (
	"github.com/spf13/cobra"
)

// Execute parses argv and runs the matching subcommand.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:           "pt",
	Short:         "Paper trading operational CLI",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(instrumentsCmd)
}
