package ptcli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print pt version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionLine())
	},
}

func versionLine() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "pt dev (no build info)"
	}
	v := bi.Main.Version
	if v == "" || v == "(devel)" {
		v = "devel"
	}
	return fmt.Sprintf("pt %s (%s)", v, bi.Main.Path)
}
