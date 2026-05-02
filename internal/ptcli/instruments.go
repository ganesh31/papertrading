package ptcli

import (
	"github.com/spf13/cobra"
)

var instrumentsCmd = &cobra.Command{
	Use:   "instruments",
	Short: "Contract master (ref.instruments) tooling",
}

var instrumentsSyncCmd = &cobra.Command{
	Use:          "sync",
	Short:        "Upsert Angel OpenAPIScripMaster JSON into ref.instruments",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ErrInstrumentsSyncNotImplemented
	},
}

func init() {
	instrumentsSyncCmd.Flags().Bool("force", false, "Re-sync even when metadata is considered fresh (reserved for P1-T03)")
	instrumentsCmd.AddCommand(instrumentsSyncCmd)
}
