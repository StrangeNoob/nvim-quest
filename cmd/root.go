package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nvim-quest/internal/lessons"
	"nvim-quest/internal/progress"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim basics through interactive terminal quests",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func lessonLoader() lessons.Loader {
	return lessons.NewLoader("lessons")
}

func progressData() (progress.Store, progress.Model, error) {
	store, err := progress.DefaultStore()
	if err != nil {
		return progress.Store{}, progress.Model{}, err
	}
	data, err := store.Load()
	return store, data, err
}

func printError(err error) error {
	return fmt.Errorf("nvim-quest: %w", err)
}
