package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nvim-quest/internal/progress"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Erase all progress and start the journey anew",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("This erases ALL progress. Type 'yes' to confirm: ")
		var answer string
		// Any read failure (empty line, EOF) leaves answer empty and aborts —
		// fail safe: never delete progress unless the user explicitly typed yes.
		if _, err := fmt.Scanln(&answer); err != nil || answer != "yes" {
			fmt.Println("aborted")
			return nil
		}
		if err := os.Remove(progress.DefaultPath()); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Println("progress erased — the journey begins again")
		return nil
	},
}

func init() { rootCmd.AddCommand(resetCmd) }
