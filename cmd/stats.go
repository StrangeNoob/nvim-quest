package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show saved quest progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, saved, err := progressData()
		if err != nil {
			return printError(err)
		}
		badges := "None yet"
		if len(saved.Badges) > 0 {
			badges = strings.Join(saved.Badges, ", ")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "XP: %d\nStreak: %d\nCompleted challenges: %d\nBadges: %s\nLast lesson: %s\n",
			saved.XP, saved.Streak, len(saved.CompletedChallenges), badges, valueOrNone(saved.LastLessonID))
		return nil
	},
}

func valueOrNone(value string) string {
	if value == "" {
		return "None yet"
	}
	return value
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
