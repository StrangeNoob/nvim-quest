package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"nvim-quest/internal/lessons"
	"nvim-quest/internal/progress"
)

func Run(selected []lessons.Lesson, saved progress.Model, store progress.Store, skipCompleted bool) error {
	program := tea.NewProgram(NewModel(selected, saved, store, skipCompleted))
	_, err := program.Run()
	return err
}
