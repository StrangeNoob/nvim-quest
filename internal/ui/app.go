// Package ui is the Bubble Tea front end: one root model, five screens.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/StrangeNoob/nvim-quest/internal/content"
	"github.com/StrangeNoob/nvim-quest/internal/engine"
	"github.com/StrangeNoob/nvim-quest/internal/progress"
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

	// release / update notice
	version      string        // current build version, shown on the title screen
	checkUpdate  func() string // off-render check; returns a newer version or "" (nil = disabled)
	updateLatest string        // a newer release to advertise, once the check returns

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
	heartMsg      string // why the last heart was lost; cleared on the next valid key
	actHeartsLost map[int]bool
	// results
	resStars        int
	resXP           int
	resBadges       []string
	resFailed       bool
	resWasBoss      bool
	resGameComplete bool // final boss cleared → show the finale celebration
}

// New builds the root model. version is the current build (shown on the title
// screen); checkUpdate is run once off the render path to learn whether a newer
// release exists — it returns the newer version string or "" (pass nil to
// disable the check entirely, e.g. for dev builds or in tests).
func New(lessons []content.Lesson, prog *progress.Progress, savePath, version string, checkUpdate func() string) Model {
	m := Model{
		width: 80, height: 24,
		lessons: lessons, prog: prog, savePath: savePath,
		version: version, checkUpdate: checkUpdate,
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

// updateCheckedMsg carries the result of the off-render update check.
type updateCheckedMsg struct{ latest string }

func (m Model) Init() tea.Cmd {
	if m.checkUpdate == nil {
		return nil
	}
	check := m.checkUpdate
	return func() tea.Msg { return updateCheckedMsg{latest: check()} }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case updateCheckedMsg:
		m.updateLatest = msg.latest
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
