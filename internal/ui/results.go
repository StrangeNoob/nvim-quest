package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch normalizeKey(keyMsg) {
	case "enter", " ":
		if m.resFailed {
			return m.startBoss()
		}
		if m.resWasBoss {
			m.scr = screenMap
			m.mapIdx = min(m.lessonIdx+1, len(m.lessons)-1)
			return m, nil
		}
		l := m.lessons[m.lessonIdx]
		if m.chIdx+1 < len(l.Challenges) {
			m.chIdx++
			return m.startChallenge(), nil
		}
		if l.Boss != nil {
			return m.startBoss()
		}
		m.scr = screenMap
		m.mapIdx = min(m.lessonIdx+1, len(m.lessons)-1)
		return m, nil
	case "esc", "q":
		m.scr = screenMap
		m.mapIdx = m.lessonIdx
		return m, nil
	}
	return m, nil
}

func (m Model) viewResults() string {
	l := m.lessons[m.lessonIdx]
	pal := paletteFor(l.Act)
	var b strings.Builder
	title := "✦ ROOM CLEAR ✦"
	if m.resWasBoss {
		title = "⚔ ACT COMPLETE ⚔"
	}
	if m.resFailed {
		title = "✖ THE TIMER RAN OUT ✖"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(pal.Primary).Bold(true).Render(title) + "\n\n")
	if m.resFailed {
		// resFailed is only set on a boss timer expiry, so l.Boss is non-nil
		// here today; guard anyway so the view can never panic.
		who := "The boss"
		if l.Boss != nil {
			who = l.Boss.Name
		}
		b.WriteString(who + " survives... this time.\n\n")
		b.WriteString(dimStyle.Render("[enter] try again · [esc] world map"))
		return b.String()
	}
	if !m.resWasBoss {
		b.WriteString(strings.Repeat("⭐", m.resStars) + strings.Repeat("☆", 3-m.resStars) + "\n")
	}
	if m.resXP > 0 {
		b.WriteString(successStyle.Render(fmt.Sprintf("+%d XP", m.resXP)) + "\n")
	} else if !m.resFailed {
		b.WriteString(dimStyle.Render("already conquered — no XP, but stars can still improve") + "\n")
	}
	for _, badge := range m.resBadges {
		b.WriteString(successStyle.Render("🏅 new badge: "+badge) + "\n")
	}
	b.WriteString(fmt.Sprintf("\nlevel %d · %d XP · combo ⚡x%d\n", m.prog.Level, m.prog.XP, m.combo))
	b.WriteString("\n" + dimStyle.Render("[enter] onward · [esc] world map"))
	return b.String()
}
