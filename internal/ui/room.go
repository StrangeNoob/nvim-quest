package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StrangeNoob/nvim-quest/internal/content"
	"github.com/StrangeNoob/nvim-quest/internal/engine"
	"github.com/StrangeNoob/nvim-quest/internal/game"
)

type tickMsg struct{ gen int }
type flashDoneMsg struct{}

func tickCmd(gen int) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{gen: gen} })
}

func flashCmd() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg { return flashDoneMsg{} })
}

func newSim(ch content.Challenge) *engine.Simulator {
	s := engine.New(ch.Buffer, engine.Pos{Row: ch.Cursor[0], Col: ch.Cursor[1]})
	if len(ch.AllowedKeys) > 0 {
		allow := map[string]bool{}
		for _, k := range ch.AllowedKeys {
			allow[k] = true
		}
		s.AllowedKeys = allow
	}
	return s
}

// openLesson enters lesson i at its first incomplete challenge, or its boss
// if only the boss remains, or replays from the top if fully complete.
func (m Model) openLesson(i int) (Model, tea.Cmd) {
	m.lessonIdx = i
	l := m.lessons[i]
	m.prog.LastLesson = l.ID
	m.chIdx = -1
	for j, ch := range l.Challenges {
		if !m.prog.IsCompleted(ch.ID) {
			m.chIdx = j
			break
		}
	}
	if m.chIdx == -1 {
		if l.Boss != nil && !m.prog.IsCompleted(l.ID+":boss") {
			return m.startBoss()
		}
		m.chIdx = 0
	}
	return m.startChallenge(), nil
}

func (m Model) startChallenge() Model {
	ch := m.lessons[m.lessonIdx].Challenges[m.chIdx]
	m.scr = screenRoom
	m.inBoss = false
	m.sim = newSim(ch)
	m.keystrokes = 0
	m.hearts = 3
	m.showHint = false
	m.flash = false
	m.heartMsg = ""
	return m
}

func (m Model) startBoss() (Model, tea.Cmd) {
	boss := m.lessons[m.lessonIdx].Boss
	m.scr = screenRoom
	m.inBoss = true
	m.bossStep = 0
	m.sim = newSim(boss.Steps[0])
	m.keystrokes = 0
	m.hearts = 3
	m.showHint = false
	m.flash = false
	m.heartMsg = ""
	m.timeLeft = boss.TimeLimitSec
	m.tickGen++
	return m, tickCmd(m.tickGen)
}

func (m *Model) cur() content.Challenge {
	l := m.lessons[m.lessonIdx]
	if m.inBoss {
		return l.Boss.Steps[m.bossStep]
	}
	return l.Challenges[m.chIdx]
}

func (m Model) updateRoom(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if msg.gen != m.tickGen || !m.inBoss || m.scr != screenRoom || m.flash {
			return m, nil
		}
		m.timeLeft--
		if m.timeLeft <= 0 {
			m.resFailed = true
			m.resWasBoss = true
			m.resGameComplete = false
			m.resStars, m.resXP, m.resBadges = 0, 0, nil
			m.combo = 1
			m.scr = screenResults
			return m, nil
		}
		return m, tickCmd(m.tickGen)
	case flashDoneMsg:
		if m.flash {
			m.flash = false
			m.scr = screenResults
		}
		return m, nil
	case tea.KeyMsg:
		if m.flash {
			return m, nil
		}
		key := normalizeKey(msg)
		// Clear the heart-loss notice on any keystroke; an invalid key resets it below.
		m.heartMsg = ""
		if key == "esc" && m.sim.Mode == engine.ModeNormal {
			m.scr = screenMap
			m.mapIdx = m.lessonIdx
			return m, nil
		}
		if key == "?" && m.sim.Mode == engine.ModeNormal {
			m.showHint = !m.showHint
			return m, nil
		}
		if key != "esc" {
			m.keystrokes++
		}
		ev := m.sim.Press(key)
		if ev.Kind == engine.EvInvalid {
			m.hearts--
			m.combo = 1
			m.heartMsg = fmt.Sprintf("✗ %s won't work here — that key isn't part of this room (lost a heart ♥)", keyLabel(key))
			m.actHeartsLost[m.lessons[m.lessonIdx].Act] = true
			if m.hearts <= 0 {
				if m.inBoss {
					return m.startBoss()
				}
				return m.startChallenge(), nil
			}
			return m, nil
		}
		if m.cur().Goal.Met(m.sim) {
			return m.handleClear()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleClear() (tea.Model, tea.Cmd) {
	l := m.lessons[m.lessonIdx]
	if m.inBoss {
		if m.bossStep < len(l.Boss.Steps)-1 {
			m.bossStep++
			m.sim = newSim(l.Boss.Steps[m.bossStep])
			return m, nil
		}
		return m.awardBoss()
	}
	return m.awardChallenge()
}

func (m Model) awardChallenge() (tea.Model, tea.Cmd) {
	ch := m.lessons[m.lessonIdx].Challenges[m.chIdx]
	stars := game.Stars(m.keystrokes, ch.Par)
	m.resStars = stars
	m.resXP, m.resBadges = 0, nil
	m.resFailed, m.resWasBoss, m.resGameComplete = false, false, false
	if !m.prog.IsCompleted(ch.ID) {
		m.resXP = game.XP(ch.XP, m.combo)
		m.prog.XP += m.resXP
		m.prog.MarkCompleted(ch.ID)
	}
	if stars > m.prog.Stars[ch.ID] {
		m.prog.Stars[ch.ID] = stars
	}
	m.prog.Level = game.LevelForXP(m.prog.XP)
	if !m.prog.HasBadge(game.BadgeFirstSteps) {
		m.prog.AddBadge(game.BadgeFirstSteps)
		m.resBadges = append(m.resBadges, game.BadgeFirstSteps)
	}
	if game.BirdieEarned(m.prog.Stars) && !m.prog.HasBadge(game.BadgeBirdie) {
		m.prog.AddBadge(game.BadgeBirdie)
		m.resBadges = append(m.resBadges, game.BadgeBirdie)
	}
	if m.hearts == 3 {
		m.combo = game.NextCombo(m.combo)
	}
	m.save()
	m.flash = true
	return m, flashCmd()
}

func (m Model) awardBoss() (tea.Model, tea.Cmd) {
	l := m.lessons[m.lessonIdx]
	bossID := l.ID + ":boss"
	m.resStars, m.resXP, m.resBadges = 0, 0, nil
	m.resFailed = false
	m.resWasBoss = true
	m.resGameComplete = m.lessonIdx == len(m.lessons)-1 // final boss → finale
	if !m.prog.IsCompleted(bossID) {
		m.resXP = game.XP(l.Boss.XP, m.combo)
		m.prog.XP += m.resXP
		m.prog.MarkCompleted(bossID)
	}
	m.prog.Level = game.LevelForXP(m.prog.XP)
	if b := game.ActBadge(l.Act); b != "" && !m.prog.HasBadge(b) {
		m.prog.AddBadge(b)
		m.resBadges = append(m.resBadges, b)
	}
	if !m.actHeartsLost[l.Act] && !m.prog.HasBadge(game.BadgeUntouchable) {
		m.prog.AddBadge(game.BadgeUntouchable)
		m.resBadges = append(m.resBadges, game.BadgeUntouchable)
	}
	if l.Act == 3 && !m.prog.HasBadge(game.BadgeGridBreaker) {
		m.prog.AddBadge(game.BadgeGridBreaker)
		m.resBadges = append(m.resBadges, game.BadgeGridBreaker)
	}
	if m.hearts == 3 {
		m.combo = game.NextCombo(m.combo)
	}
	m.save()
	m.flash = true
	return m, flashCmd()
}

func (m Model) viewRoom() string {
	l := m.lessons[m.lessonIdx]
	pal := paletteFor(l.Act)
	ch := m.cur()
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(pal.Primary).Bold(true).
		Render(fmt.Sprintf("ACT %d · %s", l.Act, actName(l.Act))) + "\n")
	if m.inBoss {
		boss := l.Boss
		b.WriteString(dangerStyle.Render("⚔ BOSS: "+boss.Name) + "\n")
		b.WriteString(dimStyle.Render(boss.Taunt) + "\n")
		b.WriteString(fmt.Sprintf("%s %d:%02d · step %d/%d\n\n",
			timerBar(m.timeLeft, boss.TimeLimitSec, 30),
			m.timeLeft/60, m.timeLeft%60, m.bossStep+1, len(boss.Steps)))
	} else {
		b.WriteString(fmt.Sprintf("%s — room %d/%d\n", l.Title, m.chIdx+1, len(l.Challenges)))
		b.WriteString(dimStyle.Render(l.Story) + "\n\n")
	}
	if m.isFreshPlayer() && !m.inBoss {
		b.WriteString(m.renderFirstTimeHelp() + "\n")
	}
	b.WriteString(ch.Intro + "\n\n")
	b.WriteString(m.renderBuffer() + "\n\n")
	b.WriteString(m.renderHUD(ch) + "\n")
	if m.heartMsg != "" && !m.flash {
		b.WriteString(dangerStyle.Render(m.heartMsg) + "\n")
	}
	if m.showHint {
		hint := ch.Hint
		if hint == "" {
			hint = "no hint for this one — trust your training."
		}
		b.WriteString(successStyle.Render("hint: "+hint) + "\n")
	}
	if m.flash {
		b.WriteString(successStyle.Render("✦ CLEARED ✦") + "\n")
	}
	if m.saveErr != "" {
		b.WriteString(dangerStyle.Render("warning: could not save progress: "+m.saveErr) + "\n")
	}
	b.WriteString(dimStyle.Render("[?] hint · [esc] back to map"))
	return b.String()
}

// renderFirstTimeHelp is the one-time "NEW HERE?" legend shown in a room while
// the player has not yet cleared their first room. It explains the HUD symbols
// a newcomer has no other way to decode.
func (m Model) renderFirstTimeHelp() string {
	bold := lipgloss.NewStyle().Bold(true)
	rule := dimStyle.Render(strings.Repeat("─", 52))
	var b strings.Builder
	b.WriteString(dimStyle.Render("─ NEW HERE? ") + dimStyle.Render(strings.Repeat("─", 40)) + "\n")
	b.WriteString(" " + dimStyle.Render("-- NORMAL --") + " is command mode — press " + bold.Render("i") + " to type, " + bold.Render("Esc") + " to return\n")
	b.WriteString(" ♥ hearts: a wrong key costs one (3 per room)\n")
	b.WriteString(" ⚡ combo: clearing rooms cleanly multiplies your XP\n")
	b.WriteString(" par: solve in this many keys (or fewer) for ⭐⭐⭐\n")
	b.WriteString(rule)
	return b.String()
}

func (m Model) renderBuffer() string {
	var lines []string
	for i, line := range m.sim.Buffer {
		switch {
		case m.flash:
			lines = append(lines, "  "+successStyle.Render(line))
		case i == m.sim.Cursor.Row:
			col := m.sim.Cursor.Col
			r := []rune(line)
			var s string
			if col >= len(r) {
				s = string(r) + cursorStyle.Render(" ")
			} else {
				s = string(r[:col]) + cursorStyle.Render(string(r[col])) + string(r[col+1:])
			}
			lines = append(lines, "  "+s)
		default:
			lines = append(lines, "  "+line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderHUD(ch content.Challenge) string {
	mode := "-- NORMAL --"
	switch m.sim.Mode {
	case engine.ModeInsert:
		mode = "-- INSERT --"
	case engine.ModeSearch:
		mode = "/" + m.sim.SearchQuery
	}
	parts := []string{mode}
	if p := m.sim.Pending(); p != "" {
		parts = append(parts, p)
	}
	parts = append(parts,
		strings.Repeat("♥ ", m.hearts)+strings.Repeat("· ", 3-m.hearts),
		fmt.Sprintf("⚡x%d", m.combo))
	if !m.inBoss {
		parts = append(parts, fmt.Sprintf("keys %d · par %d", m.keystrokes, ch.Par))
	}
	return strings.Join(parts, "   ")
}

func timerBar(left, total, width int) string {
	filled := 0
	if total > 0 {
		filled = left * width / total
	}
	if filled > width {
		filled = width
	}
	return dangerStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", width-filled))
}
