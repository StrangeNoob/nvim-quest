package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim through an epic three-act terminal quest",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("nvim-quest v2 — game wiring lands in a later task")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
