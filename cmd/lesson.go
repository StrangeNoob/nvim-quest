package cmd

import (
	"github.com/spf13/cobra"

	"nvim-quest/internal/lessons"
	"nvim-quest/internal/ui"
)

var lessonCmd = &cobra.Command{
	Use:   "lesson <lesson-id>",
	Short: "Start a specific lesson",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lesson, err := lessonLoader().ByID(args[0])
		if err != nil {
			return printError(err)
		}
		store, saved, err := progressData()
		if err != nil {
			return printError(err)
		}
		return ui.Run([]lessons.Lesson{lesson}, saved, store, false)
	},
}

func init() {
	rootCmd.AddCommand(lessonCmd)
}
