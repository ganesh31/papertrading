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
	RunE:         runInstrumentsSync,
}

func init() {
	instrumentsSyncCmd.Flags().String("file", "", "Path to OpenAPIScripMaster.json (default: download from Angel public CDN)")
	instrumentsSyncCmd.Flags().String("database-url", "", "Postgres URL (default: $DATABASE_URL)")
	instrumentsSyncCmd.Flags().Bool("force", false, "Reserved: bypass freshness / TTL checks when implemented")
	instrumentsCmd.AddCommand(instrumentsSyncCmd)
}
