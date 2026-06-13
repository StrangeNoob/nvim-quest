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
		fmt.Scanln(&answer)
		if answer != "yes" {
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
