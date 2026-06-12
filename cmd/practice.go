package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"nvim-quest/internal/ui"
)

var practiceCmd = &cobra.Command{
	Use:       "practice <topic>",
	Short:     "Practice a Vim topic",
	ValidArgs: []string{"motions", "insert", "delete", "yank-paste", "search"},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		for _, topic := range cmd.ValidArgs {
			if args[0] == topic {
				return nil
			}
		}
		return fmt.Errorf("unknown topic %q; choose motions, insert, delete, yank-paste, or search", args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		selected, err := lessonLoader().ByTopic(args[0])
		if err != nil {
			return printError(err)
		}
		store, saved, err := progressData()
		if err != nil {
			return printError(err)
		}
		return ui.Run(selected, saved, store, false)
	},
}

func init() {
	rootCmd.AddCommand(practiceCmd)
}
