package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var menuItems = []string{"Continue", "World Map", "Stats", "Quit"}

func (m Model) updateTitle(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch normalizeKey(keyMsg) {
	case "j":
		if m.menuIdx < len(menuItems)-1 {
			m.menuIdx++
		}
	case "k":
		if m.menuIdx > 0 {
			m.menuIdx--
		}
	case "q":
		return m, tea.Quit
	case "enter":
		switch menuItems[m.menuIdx] {
		case "Continue":
			return m.openLesson(m.firstIncomplete())
		case "World Map":
			m.scr = screenMap
			m.mapIdx = m.firstIncomplete()
		case "Stats":
			m.scr = screenStats
		case "Quit":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewTitle() string {
	act := m.lessons[m.firstIncomplete()].Act
	pal := paletteFor(act)
	logo := strings.Join([]string{
		`             _                                       _   `,
		` _ ____   __(_)_ __ ___        __ _ _   _  ___  ___| |_ `,
		`| '_ \ \ / /| | '_ ` + "`" + ` _ \ _____ / _` + "`" + ` | | | |/ _ \/ __| __|`,
		`| | | \ V / | | | | | | |_____| (_| | |_| |  __/\__ \ |_ `,
		`|_| |_|\_/  |_|_| |_| |_|      \__, |\__,_|\___||___/\__|`,
		`                                  |_|                    `,
	}, "\n")
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(pal.Primary).Render(logo) + "\n")
	b.WriteString(dimStyle.Render("learn the blade cursor · three acts · one journey") + "\n\n")
	for i, item := range menuItems {
		line := "  " + item
		style := lipgloss.NewStyle()
		if i == m.menuIdx {
			line = "> " + item
			style = style.Bold(true).Foreground(pal.Primary)
		}
		b.WriteString(style.Render(line) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("level %d · %d XP", m.prog.Level, m.prog.XP)))
	if m.updateLatest != "" {
		b.WriteString("\n" + successStyle.Render(
			fmt.Sprintf("✨ %s available — run `nvim-quest update`", m.updateLatest)))
	}
	return b.String()
}

func (m Model) updateStats(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch normalizeKey(keyMsg) {
		case "esc", "q", "enter":
			m.scr = screenTitle
		}
	}
	return m, nil
}

func (m Model) viewStats() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("YOUR LEGEND") + "\n\n")
	b.WriteString(fmt.Sprintf("level %d · %d XP\n", m.prog.Level, m.prog.XP))
	stars := 0
	for _, s := range m.prog.Stars {
		stars += s
	}
	b.WriteString(fmt.Sprintf("⭐ %d stars · %d rooms cleared\n", stars, len(m.prog.Completed)))
	if len(m.prog.Badges) > 0 {
		b.WriteString("🏅 " + strings.Join(m.prog.Badges, " · ") + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("[esc] back"))
	return b.String()
}
