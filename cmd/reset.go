package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var resetYes bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all local progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetYes {
			fmt.Fprint(cmd.OutOrStdout(), "Reset all nvim-quest progress? [y/N] ")
			answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil {
				return printError(err)
			}
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Fprintln(cmd.OutOrStdout(), "Reset cancelled.")
				return nil
			}
		}
		store, _, err := progressData()
		if err != nil {
			return printError(err)
		}
		if err := store.Reset(); err != nil {
			return printError(err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Progress reset.")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&resetYes, "yes", "y", false, "reset without confirmation")
	rootCmd.AddCommand(resetCmd)
}
