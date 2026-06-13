package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/StrangeNoob/nvim-quest/internal/progress"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your journey stats",
	Run: func(cmd *cobra.Command, args []string) {
		p := progress.Load(progress.DefaultPath())
		stars := 0
		for _, s := range p.Stars {
			stars += s
		}
		fmt.Printf("level %d · %d XP\n", p.Level, p.XP)
		fmt.Printf("rooms cleared: %d · stars: %d\n", len(p.Completed), stars)
		if len(p.Badges) > 0 {
			fmt.Println("badges: " + strings.Join(p.Badges, ", "))
		}
		if p.LastLesson != "" {
			fmt.Println("last lesson: " + p.LastLesson)
		}
	},
}

func init() { rootCmd.AddCommand(statsCmd) }
