package ui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"nvim-quest/internal/content"
	"nvim-quest/internal/progress"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func press(t *testing.T, m Model, keys ...string) Model {
	t.Helper()
	for _, k := range keys {
		mm, _ := m.Update(key(k))
		m = mm.(Model)
	}
	return m
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	lessons, err := content.All()
	if err != nil {
		t.Fatal(err)
	}
	return New(lessons, progress.New(), filepath.Join(t.TempDir(), "progress.json"))
}

func TestWelcomeScreen(t *testing.T) {
	// A fresh player (no rooms cleared) meets the welcome screen first.
	m := newTestModel(t)
	if m.scr != screenWelcome {
		t.Fatalf("fresh player initial screen = %v, want welcome", m.scr)
	}
	m = press(t, m, "enter")
	if m.scr != screenTitle {
		t.Fatalf("after enter screen = %v, want title", m.scr)
	}

	// A returning player (has cleared a room) skips straight to the title.
	lessons, err := content.All()
	if err != nil {
		t.Fatal(err)
	}
	prog := progress.New()
	prog.MarkCompleted("a1l1c1")
	r := New(lessons, prog, t.TempDir())
	if r.scr != screenTitle {
		t.Fatalf("returning player initial screen = %v, want title", r.scr)
	}
}

func TestTitleNavigation(t *testing.T) {
	m := newTestModel(t)
	m.scr = screenTitle // past the first-launch welcome
	if m.scr != screenTitle {
		t.Fatalf("initial screen = %v, want title", m.scr)
	}
	m = press(t, m, "j", "enter") // menu item 1: World Map
	if m.scr != screenMap {
		t.Fatalf("screen = %v, want map", m.scr)
	}
	m = press(t, m, "esc")
	if m.scr != screenTitle {
		t.Fatalf("screen = %v, want title", m.scr)
	}
	m = press(t, m, "j", "enter") // menuIdx 1 -> 2: Stats
	if m.scr != screenStats {
		t.Fatalf("screen = %v, want stats", m.scr)
	}
	m = press(t, m, "esc")
	if m.scr != screenTitle {
		t.Fatalf("screen = %v, want title", m.scr)
	}
}

func TestArrowKeysDoNotNavigate(t *testing.T) {
	// This is a Vim trainer: only j/k move, never the arrow keys.
	m := newTestModel(t)
	m.scr = screenTitle // past the first-launch welcome
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mm.(Model)
	if m.menuIdx != 0 {
		t.Errorf("title menuIdx = %d after Down arrow, want 0 (arrows must not move)", m.menuIdx)
	}

	m.scr = screenMap
	m.mapIdx = 0
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mm.(Model)
	if m.mapIdx != 0 {
		t.Errorf("map mapIdx = %d after Down arrow, want 0 (arrows must not move)", m.mapIdx)
	}
}

func TestMapLockAndEnter(t *testing.T) {
	m := newTestModel(t)
	m.scr = screenMap
	if !m.unlocked(0) {
		t.Fatal("lesson 0 must be unlocked")
	}
	if m.unlocked(1) {
		t.Fatal("lesson 1 must be locked at fresh start")
	}
	m = press(t, m, "j", "enter") // locked lesson: must stay on map
	if m.scr != screenMap {
		t.Fatalf("locked lesson must not open, screen = %v", m.scr)
	}
	m = press(t, m, "k", "enter")
	if m.scr != screenRoom {
		t.Fatalf("screen = %v, want room", m.scr)
	}
}

func TestPlayThroughFirstChallenge(t *testing.T) {
	m := newTestModel(t)
	m, _ = m.openLesson(0)
	// a1l1c1: insert "I " at the front of "am ready" (par 3)
	m = press(t, m, "i", "I", " ")
	if !m.flash {
		t.Fatal("expected success flash after goal met")
	}
	mm, _ := m.Update(flashDoneMsg{})
	m = mm.(Model)
	if m.scr != screenResults {
		t.Fatalf("screen = %v, want results", m.scr)
	}
	if m.resStars != 3 {
		t.Errorf("stars = %d, want 3 (3 keys vs par 3)", m.resStars)
	}
	if m.resXP != 50 {
		t.Errorf("xp = %d, want 50 (base 50 x combo 1)", m.resXP)
	}
	if !m.prog.IsCompleted("a1l1c1") {
		t.Error("challenge not marked complete")
	}
	if m.combo != 2 {
		t.Errorf("combo = %d, want 2 after clean clear", m.combo)
	}
	m = press(t, m, "enter")
	if m.scr != screenRoom || m.chIdx != 1 {
		t.Fatalf("want next room (chIdx 1), got screen=%v chIdx=%d", m.scr, m.chIdx)
	}
}

func TestInvalidKeyCostsHeartAndResetsRoom(t *testing.T) {
	m := newTestModel(t)
	m, _ = m.openLesson(0)
	m.combo = 3
	m = press(t, m, "z") // only "i" is allowed in a1l1c1
	if m.hearts != 2 {
		t.Errorf("hearts = %d, want 2", m.hearts)
	}
	if m.combo != 1 {
		t.Errorf("combo = %d, want reset to 1", m.combo)
	}
	m = press(t, m, "z", "z") // two more mistakes: hearts hit 0, room resets
	if m.hearts != 3 {
		t.Errorf("hearts = %d, want 3 after room reset", m.hearts)
	}
	if m.keystrokes != 0 {
		t.Errorf("keystrokes = %d, want 0 after reset", m.keystrokes)
	}
}

func TestBossTimerFailureAndRetry(t *testing.T) {
	m := newTestModel(t)
	m.lessonIdx = 3 // act1-04-the-great-leaps carries Sensei's Trial
	m, _ = m.startBoss()
	if !m.inBoss || m.timeLeft != 60 {
		t.Fatalf("boss not started: inBoss=%v timeLeft=%d", m.inBoss, m.timeLeft)
	}
	for i := 0; i < 60; i++ {
		mm, _ := m.Update(tickMsg{gen: m.tickGen})
		m = mm.(Model)
	}
	if m.scr != screenResults || !m.resFailed {
		t.Fatalf("expected failed results, screen=%v failed=%v", m.scr, m.resFailed)
	}
	m = press(t, m, "enter") // retry
	if m.scr != screenRoom || !m.inBoss || m.timeLeft != 60 || m.bossStep != 0 {
		t.Fatal("retry must restart the boss from step 1 with a fresh timer")
	}
}

func TestStaleTickIsIgnored(t *testing.T) {
	m := newTestModel(t)
	m.lessonIdx = 3
	m, _ = m.startBoss()
	mm, _ := m.Update(tickMsg{gen: m.tickGen - 1}) // stale generation
	m = mm.(Model)
	if m.timeLeft != 60 {
		t.Errorf("stale tick must not drain timer, timeLeft = %d", m.timeLeft)
	}
}
