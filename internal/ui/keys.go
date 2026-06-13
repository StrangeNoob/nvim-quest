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

// keyLabel renders an engine key for display in player-facing messages.
func keyLabel(key string) string {
	switch key {
	case " ":
		return "space"
	case "enter":
		return "enter"
	default:
		return "'" + key + "'"
	}
}
