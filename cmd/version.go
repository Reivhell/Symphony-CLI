package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time metadata (overridden via -ldflags -X).
var (
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Symphony CLI version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Symphony CLI")
		fmt.Printf("  Version   : %s\n", Version)
		fmt.Printf("  Commit    : %s\n", Commit)
		fmt.Printf("  Built     : %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
