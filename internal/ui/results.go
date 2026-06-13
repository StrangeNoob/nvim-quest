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
	if m.resGameComplete {
		return m.viewFinale()
	}
	l := m.lessons[m.lessonIdx]
	pal := paletteFor(l.Act)
	var b strings.Builder
	title := "✦ ROOM CLEAR ✦"
	if m.resWasBoss {
		title = fmt.Sprintf("⚔ ACT %s COMPLETE ⚔", roman(l.Act))
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
	if m.resWasBoss && l.Boss != nil {
		b.WriteString(successStyle.Render(l.Boss.Name+" defeated!") + "\n")
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

	// On a boss clear, announce the act that just unlocked so progression is clear.
	if m.resWasBoss && m.lessonIdx+1 < len(m.lessons) {
		next := m.lessons[m.lessonIdx+1]
		np := paletteFor(next.Act)
		b.WriteString("\n" + dimStyle.Render(strings.Repeat("─", 40)) + "\n")
		b.WriteString(np.PrimaryStyle().Bold(true).Render(
			fmt.Sprintf("UNLOCKED ▶ ACT %s · %s", roman(next.Act), actName(next.Act))) + "\n")
		b.WriteString(dimStyle.Render("learn: "+actSummary(next.Act)) + "\n")
		b.WriteString("\n" + dimStyle.Render("[enter] enter the next act · [esc] world map"))
		return b.String()
	}
	b.WriteString("\n" + dimStyle.Render("[enter] onward · [esc] world map"))
	return b.String()
}

// viewFinale is the celebration shown after the final boss is cleared.
func (m Model) viewFinale() string {
	pal := paletteFor(3)
	stars, maxStars := 0, 0
	for _, l := range m.lessons {
		for _, ch := range l.Challenges {
			maxStars += 3
			stars += m.prog.Stars[ch.ID]
		}
	}
	var b strings.Builder
	b.WriteString(pal.PrimaryStyle().Bold(true).Render("🎉  YOU MASTERED THE BLADE CURSOR  🎉") + "\n\n")
	b.WriteString("The Grid Core is broken. The journey is complete.\n\n")
	b.WriteString(successStyle.Render(fmt.Sprintf("level %d · %d XP · ⭐ %d/%d stars", m.prog.Level, m.prog.XP, stars, maxStars)) + "\n")
	b.WriteString(fmt.Sprintf("%d rooms cleared · 🏅 %d badges\n", len(m.prog.Completed), len(m.prog.Badges)))
	if len(m.prog.Badges) > 0 {
		b.WriteString(dimStyle.Render(strings.Join(m.prog.Badges, " · ")) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("[enter] return to your journey"))
	return b.String()
}

// roman renders an act number (1-3) as a Roman numeral for the act headings.
func roman(n int) string {
	switch n {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	}
	return fmt.Sprint(n)
}
