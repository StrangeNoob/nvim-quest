// Package ui is the Bubble Tea front end: one root model, five screens.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nvim-quest/internal/content"
	"nvim-quest/internal/engine"
	"nvim-quest/internal/progress"
)

type screen int

const (
	screenTitle screen = iota
	screenMap
	screenStats
	screenRoom
	screenResults
	screenWelcome
)

type Model struct {
	width, height int
	scr           screen

	lessons  []content.Lesson
	prog     *progress.Progress
	savePath string
	saveErr  string

	// title
	menuIdx int
	// map
	mapIdx int
	// room / boss
	lessonIdx     int
	chIdx         int
	inBoss        bool
	bossStep      int
	timeLeft      int
	tickGen       int
	sim           *engine.Simulator
	keystrokes    int
	hearts        int
	combo         int
	showHint      bool
	flash         bool
	actHeartsLost map[int]bool
	// results
	resStars   int
	resXP      int
	resBadges  []string
	resFailed  bool
	resWasBoss bool
}

func New(lessons []content.Lesson, prog *progress.Progress, savePath string) Model {
	m := Model{
		width: 80, height: 24,
		lessons: lessons, prog: prog, savePath: savePath,
		combo: 1, actHeartsLost: map[int]bool{},
	}
	// A brand-new player meets the welcome screen before anything else.
	if m.isFreshPlayer() {
		m.scr = screenWelcome
	}
	return m
}

// isFreshPlayer reports whether the player has not yet cleared a single room.
// First-time guidance (the welcome screen, the in-room legend) shows only while
// this is true and disappears for good once the first room is cleared.
func (m *Model) isFreshPlayer() bool { return len(m.prog.Completed) == 0 }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	switch m.scr {
	case screenWelcome:
		return m.updateWelcome(msg)
	case screenMap:
		return m.updateMap(msg)
	case screenStats:
		return m.updateStats(msg)
	case screenRoom:
		return m.updateRoom(msg)
	case screenResults:
		return m.updateResults(msg)
	default:
		return m.updateTitle(msg)
	}
}

func (m Model) View() string {
	if m.width < 80 || m.height < 24 {
		return "Please resize your terminal to at least 80x24 to play nvim-quest."
	}
	var body string
	switch m.scr {
	case screenWelcome:
		body = m.viewWelcome()
	case screenMap:
		body = m.viewMap()
	case screenStats:
		body = m.viewStats()
	case screenRoom:
		body = m.viewRoom()
	case screenResults:
		body = m.viewResults()
	default:
		body = m.viewTitle()
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

func (m *Model) save() {
	if err := progress.Save(m.savePath, m.prog); err != nil {
		m.saveErr = err.Error()
	}
}
