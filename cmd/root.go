package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"nvim-quest/internal/content"
	"nvim-quest/internal/progress"
	"nvim-quest/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim through an epic three-act terminal quest",
	RunE: func(cmd *cobra.Command, args []string) error {
		lessons, err := content.All()
		if err != nil {
			return err
		}
		path := progress.DefaultPath()
		prog := progress.Load(path)
		p := tea.NewProgram(ui.New(lessons, prog, path), tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
