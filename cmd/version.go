package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version, commit, and build date",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nvim-quest %s (commit %s, built %s)\n", buildVersion, buildCommit, buildDate)
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
