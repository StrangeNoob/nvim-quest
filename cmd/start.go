package cmd

import (
	"github.com/spf13/cobra"

	"nvim-quest/internal/ui"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Continue the lesson path from saved progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, err := lessonLoader().All()
		if err != nil {
			return printError(err)
		}
		store, saved, err := progressData()
		if err != nil {
			return printError(err)
		}
		return ui.Run(all, saved, store, true)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
