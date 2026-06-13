package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nvim-quest/internal/content"
)

func (m *Model) lessonComplete(l content.Lesson) bool {
	for _, ch := range l.Challenges {
		if !m.prog.IsCompleted(ch.ID) {
			return false
		}
	}
	if l.Boss != nil && !m.prog.IsCompleted(l.ID+":boss") {
		return false
	}
	return true
}

func (m *Model) unlocked(i int) bool {
	if i == 0 {
		return true
	}
	return m.lessonComplete(m.lessons[i-1])
}

func (m *Model) firstIncomplete() int {
	for i, l := range m.lessons {
		if !m.lessonComplete(l) {
			return i
		}
	}
	return len(m.lessons) - 1
}

func (m Model) updateMap(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch normalizeKey(keyMsg) {
	case "j":
		if m.mapIdx < len(m.lessons)-1 {
			m.mapIdx++
		}
	case "k":
		if m.mapIdx > 0 {
			m.mapIdx--
		}
	case "esc", "q":
		m.scr = screenTitle
	case "enter":
		if m.unlocked(m.mapIdx) {
			return m.openLesson(m.mapIdx)
		}
	}
	return m, nil
}

func (m Model) viewMap() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("✨ YOUR JOURNEY ✨") + "\n")
	currentAct := 0
	for i, l := range m.lessons {
		if l.Act != currentAct {
			currentAct = l.Act
			pal := paletteFor(l.Act)
			b.WriteString("\n" + lipgloss.NewStyle().Foreground(pal.Primary).Bold(true).
				Render(fmt.Sprintf("ACT %d · %s", l.Act, actName(l.Act))) + "\n")
		}
		icon := "🔒"
		switch {
		case m.lessonComplete(l):
			icon = "✓"
		case m.unlocked(i):
			icon = "▶"
		}
		marker := "  "
		if i == m.mapIdx {
			marker = "> "
		}
		line := fmt.Sprintf("%s%s %s", marker, icon, l.Title)
		if l.Boss != nil {
			line += " · boss: " + l.Boss.Name
		}
		if s := m.lessonStars(l); s != "" {
			line += "  " + s
		}
		style := lipgloss.NewStyle()
		if i == m.mapIdx {
			style = style.Bold(true)
		}
		if !m.unlocked(i) {
			style = dimStyle
		}
		b.WriteString(style.Render(line) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("[j/k] move · [enter] play · [esc] title"))
	return b.String()
}

func (m Model) lessonStars(l content.Lesson) string {
	total, earned := 0, 0
	for _, ch := range l.Challenges {
		total += 3
		earned += m.prog.Stars[ch.ID]
	}
	if earned == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d⭐", earned, total)
}
