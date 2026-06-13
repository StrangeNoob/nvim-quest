package ui

import tea "github.com/charmbracelet/bubbletea"

// normalizeKey maps Bubble Tea key names onto the engine's vocabulary.
func normalizeKey(msg tea.KeyMsg) string {
	s := msg.String()
	if s == "space" {
		return " "
	}
	return s
}
