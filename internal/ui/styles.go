package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	sectionStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	cursorStyle   = lipgloss.NewStyle().Reverse(true)
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	progressStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
)
