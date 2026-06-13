package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch normalizeKey(keyMsg) {
		case "enter", " ", "esc":
			m.scr = screenTitle
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewWelcome() string {
	bold := lipgloss.NewStyle().Bold(true)
	title := lipgloss.NewStyle().Foreground(paletteFor(1).Primary).Bold(true)
	var b strings.Builder
	b.WriteString(title.Render("✨ WELCOME, TRAVELLER ✨") + "\n\n")
	b.WriteString("You'll learn real Neovim by playing it — you press the\n")
	b.WriteString("actual keys (" + bold.Render("h j k l") + ", not the arrow keys).\n\n")
	b.WriteString("Your journey runs through three worlds:\n")
	b.WriteString(lipgloss.NewStyle().Foreground(paletteFor(1).Primary).Render("  Act I   · The Cursor Dojo") +
		dimStyle.Render("   — move and edit") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(paletteFor(2).Primary).Render("  Act II  · The Motion Crypts") +
		dimStyle.Render(" — delete, change, yank") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(paletteFor(3).Primary).Render("  Act III · The Neon Grid") +
		dimStyle.Render("     — search and counts") + "\n\n")
	b.WriteString(bold.Render("On the map") + "\n")
	b.WriteString("  ▶  where you are now\n")
	b.WriteString("  🔒 locked — clear the lesson before it to open\n")
	b.WriteString("  ⭐ stars — solve a room in few keystrokes to earn 3\n\n")
	b.WriteString(bold.Render("In a room") + "\n")
	b.WriteString("  -- NORMAL -- is command mode; press " + bold.Render("i") + " to type, " + bold.Render("Esc") + " to leave\n")
	b.WriteString("  ♥♥♥ hearts — a wrong key costs one\n")
	b.WriteString("  ⚡ combo — clean clears multiply your XP; press " + bold.Render("?") + " for a hint\n\n")
	b.WriteString(title.Render("[enter] begin your journey"))
	return b.String()
}
