# nvim-quest v2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Full rewrite of nvim-quest into a real-time-keystroke Vim game with a three-act world map, par/star scoring, combos, hearts, and timed boss fights.

**Architecture:** A pure-Go Vim emulator (`internal/engine`) exposes `Press(key) → Event`; a Bubble Tea root model (`internal/ui`) switches between five screens (title, map, room, results, stats) and feeds keystrokes to the engine; lesson content is JSON embedded via `go:embed` (`assets/` + `internal/content`); mechanics math lives in `internal/game`; persistence in `internal/progress` (versioned JSON v2).

**Tech Stack:** Go 1.24, bubbletea, lipgloss, cobra. Tests with the standard library (`go test ./...`).

**Spec:** `docs/superpowers/specs/2026-06-13-nvim-quest-v2-design.md`

---

## File structure

```
main.go                          entry point (unchanged shape: calls cmd.Execute)
cmd/root.go                      cobra root: launches the TUI game
cmd/stats.go                     prints progress summary
cmd/reset.go                     confirm + wipe progress
assets/assets.go                 go:embed of lessons/*.json
assets/lessons/*.json            13 lesson files (10 lessons, 3 carry boss blocks)
internal/engine/engine.go        Simulator, modes, Press dispatch, motions, helpers
internal/engine/edit.go          insert mode, x, operators (dw/dd/cw/cc), yank/paste, undo
internal/engine/search.go        search mode, n, jumpToMatch
internal/engine/goal.go          Goal type + Met() validators
internal/engine/*_test.go        table-driven keystroke tests
internal/content/model.go        Lesson, Challenge, Boss structs
internal/content/loader.go       All() from embedded FS
internal/content/loader_test.go  content-integrity test
internal/game/game.go            Stars, NextCombo, XP, LevelForXP, badge constants
internal/game/game_test.go
internal/progress/progress.go    Progress v2, Load/Save/DefaultPath, v1 backup
internal/progress/progress_test.go
internal/ui/keys.go              normalizeKey
internal/ui/styles.go            per-act palettes
internal/ui/app.go               root Model, screen enum, Update/View dispatch
internal/ui/title.go             title + stats screens
internal/ui/worldmap.go          map screen, unlock logic
internal/ui/room.go              gameplay: challenge + boss, HUD, completion/award
internal/ui/results.go           results screen
internal/ui/app_test.go          screen-transition tests
```

Old v1 packages (`internal/engine`, `internal/lessons`, `internal/progress`, `internal/scoring`, `internal/ui`, `cmd`, `lessons/`) are deleted in Task 1 and rebuilt fresh.

---

### Task 1: Demolish v1 and scaffold the new skeleton

**Files:**
- Delete: `cmd/`, `internal/`, `lessons/`, `main.go`
- Create: `main.go`, `cmd/root.go`, `assets/assets.go`, `assets/lessons/.gitkeep`

- [ ] **Step 1: Remove v1 code**

```bash
git rm -r -q cmd internal lessons main.go
```

- [ ] **Step 2: Create `main.go`**

```go
package main

import "nvim-quest/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 3: Create `cmd/root.go` (placeholder run, real wiring in Task 12)**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim through an epic three-act terminal quest",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("nvim-quest v2 — game wiring lands in a later task")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create `assets/assets.go`**

```go
// Package assets embeds the game's lesson content.
package assets

import "embed"

//go:embed lessons/*.json
var Lessons embed.FS
```

Also create an empty placeholder so the embed pattern matches before Task 8 adds real lessons:

```bash
mkdir -p assets/lessons
printf '{"id":"placeholder","act":1,"order":999,"title":"placeholder","story":"","challenges":[]}\n' > assets/lessons/placeholder.json
```

(The placeholder is deleted in Task 8 when real lessons land.)

- [ ] **Step 5: Tidy deps and verify build**

```bash
go mod tidy && go build ./...
```

Expected: builds cleanly; `go.sum` drops bubbles if unused (bubbletea/lipgloss/cobra remain or are re-added by later tasks).

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "chore: demolish v1, scaffold v2 skeleton with embedded assets"
```

---

### Task 2: Engine core — Simulator, events, normal-mode motions

**Files:**
- Create: `internal/engine/engine.go`
- Test: `internal/engine/engine_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import "testing"

func TestMotions(t *testing.T) {
	tests := []struct {
		name   string
		buffer []string
		cursor Pos
		keys   []string
		want   Pos
	}{
		{"h moves left", []string{"abc"}, Pos{0, 1}, []string{"h"}, Pos{0, 0}},
		{"h stops at col 0", []string{"abc"}, Pos{0, 0}, []string{"h"}, Pos{0, 0}},
		{"l moves right", []string{"abc"}, Pos{0, 0}, []string{"l"}, Pos{0, 1}},
		{"l stops at line end", []string{"ab"}, Pos{0, 1}, []string{"l"}, Pos{0, 1}},
		{"j moves down and clamps col", []string{"abcdef", "ab"}, Pos{0, 5}, []string{"j"}, Pos{1, 1}},
		{"j stops at last row", []string{"a", "b"}, Pos{1, 0}, []string{"j"}, Pos{1, 0}},
		{"k moves up", []string{"a", "b"}, Pos{1, 0}, []string{"k"}, Pos{0, 0}},
		{"0 jumps to line start", []string{"hello"}, Pos{0, 4}, []string{"0"}, Pos{0, 0}},
		{"$ jumps to line end", []string{"hello"}, Pos{0, 0}, []string{"$"}, Pos{0, 4}},
		{"gg jumps to top", []string{"a", "b", "c"}, Pos{2, 0}, []string{"g", "g"}, Pos{0, 0}},
		{"G jumps to bottom", []string{"a", "b", "c"}, Pos{0, 0}, []string{"G"}, Pos{2, 0}},
		{"w jumps to next word", []string{"foo bar baz"}, Pos{0, 0}, []string{"w"}, Pos{0, 4}},
		{"w crosses to next line", []string{"foo", "bar"}, Pos{0, 0}, []string{"w"}, Pos{1, 0}},
		{"w at last word stays on line end", []string{"foo bar"}, Pos{0, 4}, []string{"w"}, Pos{0, 6}},
		{"b jumps to prev word start", []string{"foo bar"}, Pos{0, 4}, []string{"b"}, Pos{0, 0}},
		{"b inside word jumps to its start", []string{"foo bar"}, Pos{0, 6}, []string{"b"}, Pos{0, 4}},
		{"b crosses to prev line", []string{"foo", "bar"}, Pos{1, 0}, []string{"b"}, Pos{0, 0}},
		{"count 3w", []string{"a b c d"}, Pos{0, 0}, []string{"3", "w"}, Pos{0, 6}},
		{"count 2j", []string{"a", "b", "c"}, Pos{0, 0}, []string{"2", "j"}, Pos{2, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if s.Cursor != tt.want {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.want)
			}
		})
	}
}

func TestNewCopiesBuffer(t *testing.T) {
	src := []string{"abc"}
	s := New(src, Pos{0, 0})
	src[0] = "mutated"
	if s.Buffer[0] != "abc" {
		t.Errorf("New must copy the buffer, got %q", s.Buffer[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run 'TestMotions|TestNewCopies' -v`
Expected: FAIL (package doesn't compile: `Pos`, `New`, `Press` undefined)

- [ ] **Step 3: Write the implementation — `internal/engine/engine.go`**

```go
// Package engine is a small real-time Vim emulator: feed it keystrokes,
// it mutates buffer/cursor/mode state and reports what happened.
package engine

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeSearch
)

type EventKind int

const (
	EvNone EventKind = iota
	EvMoved
	EvEdited
	EvModeChanged
	EvPending
	EvSearchJumped
	EvInvalid
)

type Event struct {
	Kind EventKind
}

type Pos struct{ Row, Col int }

type snapshot struct {
	buffer []string
	cursor Pos
}

type Simulator struct {
	Buffer []string
	Cursor Pos
	Mode   Mode

	// AllowedKeys restricts normal-mode commands; nil allows everything.
	AllowedKeys map[string]bool

	pendingOp    string
	opCount      int
	pendingCount int
	pendingG     bool

	yank []string
	undo []snapshot

	SearchQuery string
	lastSearch  string
}

func New(lines []string, cursor Pos) *Simulator {
	buf := make([]string, len(lines))
	copy(buf, lines)
	return &Simulator{Buffer: buf, Cursor: cursor}
}

// Pending returns the visible pending state ("d", "3", "g") for the HUD.
func (s *Simulator) Pending() string {
	out := ""
	if s.pendingCount > 0 {
		out += itoa(s.pendingCount)
	}
	if s.pendingOp != "" {
		out += s.pendingOp
	}
	if s.pendingG {
		out += "g"
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func (s *Simulator) LastSearch() string { return s.lastSearch }

func (s *Simulator) Press(key string) Event {
	switch s.Mode {
	case ModeInsert:
		return s.pressInsert(key)
	case ModeSearch:
		return s.pressSearch(key)
	default:
		return s.pressNormal(key)
	}
}

func (s *Simulator) pressNormal(key string) Event {
	if key == "esc" {
		s.clearPending()
		return Event{EvNone}
	}
	if s.AllowedKeys != nil && !s.AllowedKeys[key] {
		s.clearPending()
		return Event{EvInvalid}
	}
	if s.pendingG {
		s.pendingG = false
		if key == "g" {
			s.Cursor = Pos{0, 0}
			return Event{EvMoved}
		}
		return Event{EvInvalid}
	}
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' && !(key == "0" && s.pendingCount == 0) {
		s.pendingCount = s.pendingCount*10 + int(key[0]-'0')
		return Event{EvPending}
	}
	if s.pendingOp != "" {
		return s.applyOperator(key)
	}
	count := s.takeCount()
	switch key {
	case "h", "j", "k", "l", "w", "b":
		for i := 0; i < count; i++ {
			s.move(key)
		}
		return Event{EvMoved}
	case "0":
		s.Cursor.Col = 0
		return Event{EvMoved}
	case "$":
		s.Cursor.Col = max(0, len(s.line())-1)
		return Event{EvMoved}
	case "G":
		s.Cursor = Pos{len(s.Buffer) - 1, 0}
		return Event{EvMoved}
	case "g":
		s.pendingG = true
		return Event{EvPending}
	case "d", "c", "y":
		s.pendingOp = key
		s.opCount = count
		return Event{EvPending}
	case "x":
		s.snapshot()
		for i := 0; i < count; i++ {
			s.deleteChar()
		}
		return Event{EvEdited}
	case "i":
		s.snapshot()
		s.Mode = ModeInsert
		return Event{EvModeChanged}
	case "a":
		s.snapshot()
		if len(s.line()) > 0 {
			s.Cursor.Col++
		}
		s.Mode = ModeInsert
		return Event{EvModeChanged}
	case "u":
		return s.applyUndo()
	case "p":
		return s.paste()
	case "/":
		s.Mode = ModeSearch
		s.SearchQuery = ""
		return Event{EvModeChanged}
	case "n":
		if s.jumpToMatch(s.lastSearch, true) {
			return Event{EvSearchJumped}
		}
		return Event{EvNone}
	}
	return Event{EvInvalid}
}

func (s *Simulator) clearPending() {
	s.pendingOp = ""
	s.opCount = 0
	s.pendingCount = 0
	s.pendingG = false
}

func (s *Simulator) takeCount() int {
	c := s.pendingCount
	s.pendingCount = 0
	if c == 0 {
		return 1
	}
	return c
}

func (s *Simulator) line() string { return s.Buffer[s.Cursor.Row] }

func (s *Simulator) setLine(text string) { s.Buffer[s.Cursor.Row] = text }

func (s *Simulator) clampCol() {
	if s.Cursor.Col >= len(s.line()) {
		s.Cursor.Col = max(0, len(s.line())-1)
	}
}

func (s *Simulator) move(key string) {
	switch key {
	case "h":
		if s.Cursor.Col > 0 {
			s.Cursor.Col--
		}
	case "l":
		if s.Cursor.Col < len(s.line())-1 {
			s.Cursor.Col++
		}
	case "j":
		if s.Cursor.Row < len(s.Buffer)-1 {
			s.Cursor.Row++
			s.clampCol()
		}
	case "k":
		if s.Cursor.Row > 0 {
			s.Cursor.Row--
			s.clampCol()
		}
	case "w":
		s.wordForward()
	case "b":
		s.wordBack()
	}
}

func (s *Simulator) wordForward() {
	line := s.line()
	col := s.Cursor.Col
	for col < len(line) && line[col] != ' ' {
		col++
	}
	for col < len(line) && line[col] == ' ' {
		col++
	}
	if col < len(line) {
		s.Cursor.Col = col
		return
	}
	if s.Cursor.Row < len(s.Buffer)-1 {
		s.Cursor.Row++
		next := s.line()
		c := 0
		for c < len(next) && next[c] == ' ' {
			c++
		}
		s.Cursor.Col = min(c, max(0, len(next)-1))
		return
	}
	s.Cursor.Col = max(0, len(line)-1)
}

func (s *Simulator) wordBack() {
	line := s.line()
	col := s.Cursor.Col
	if col == 0 {
		if s.Cursor.Row == 0 {
			return
		}
		s.Cursor.Row--
		line = s.line()
		col = len(line)
	}
	col--
	for col > 0 && col < len(line) && line[col] == ' ' {
		col--
	}
	for col > 0 && line[col-1] != ' ' {
		col--
	}
	s.Cursor.Col = max(0, col)
}
```

Add temporary stubs at the bottom of `engine.go` so this task compiles alone (replaced by real code in Tasks 3–5):

```go
// Stubs — implemented in later tasks.
func (s *Simulator) pressInsert(key string) Event  { return Event{EvNone} }
func (s *Simulator) pressSearch(key string) Event  { return Event{EvNone} }
func (s *Simulator) applyOperator(key string) Event { s.pendingOp = ""; return Event{EvInvalid} }
func (s *Simulator) deleteChar()                   {}
func (s *Simulator) snapshot()                     {}
func (s *Simulator) applyUndo() Event              { return Event{EvNone} }
func (s *Simulator) paste() Event                  { return Event{EvNone} }
func (s *Simulator) jumpToMatch(term string, fromNext bool) bool { return false }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/engine/ -v`
Expected: PASS (all motion subtests)

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): simulator core with normal-mode motions and counts"
```

---

### Task 3: Engine — insert mode

**Files:**
- Create: `internal/engine/edit.go` (move `pressInsert` stub here, make it real)
- Modify: `internal/engine/engine.go` (delete the `pressInsert` stub)
- Test: `internal/engine/edit_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
	"slices"
	"testing"
)

func TestInsertMode(t *testing.T) {
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantBuffer []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"i inserts before cursor", []string{"am ready"}, Pos{0, 0},
			[]string{"i", "I", " "}, []string{"I am ready"}, Pos{0, 2}, ModeInsert},
		{"a appends after cursor", []string{"ab"}, Pos{0, 0},
			[]string{"a", "X"}, []string{"aXb"}, Pos{0, 2}, ModeInsert},
		{"a at line end appends at end", []string{"ab"}, Pos{0, 1},
			[]string{"a", "c"}, []string{"abc"}, Pos{0, 3}, ModeInsert},
		{"esc returns to normal and steps back", []string{"ab"}, Pos{0, 0},
			[]string{"i", "X", "esc"}, []string{"Xab"}, Pos{0, 0}, ModeNormal},
		{"backspace deletes prev char", []string{"ab"}, Pos{0, 1},
			[]string{"i", "X", "backspace"}, []string{"ab"}, Pos{0, 1}, ModeInsert},
		{"a on empty line inserts", []string{""}, Pos{0, 0},
			[]string{"a", "z"}, []string{"z"}, Pos{0, 1}, ModeInsert},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if !slices.Equal(s.Buffer, tt.wantBuffer) {
				t.Errorf("buffer = %q, want %q", s.Buffer, tt.wantBuffer)
			}
			if s.Cursor != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.wantCursor)
			}
			if s.Mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", s.Mode, tt.wantMode)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run TestInsertMode -v`
Expected: FAIL (stub does nothing — buffers unchanged)

- [ ] **Step 3: Implement — create `internal/engine/edit.go`, delete the `pressInsert` stub from `engine.go`**

```go
package engine

func (s *Simulator) pressInsert(key string) Event {
	switch key {
	case "esc":
		s.Mode = ModeNormal
		if s.Cursor.Col > 0 {
			s.Cursor.Col--
		}
		s.clampCol()
		return Event{EvModeChanged}
	case "backspace":
		line := s.line()
		if s.Cursor.Col > 0 && s.Cursor.Col <= len(line) {
			s.setLine(line[:s.Cursor.Col-1] + line[s.Cursor.Col:])
			s.Cursor.Col--
		}
		return Event{EvEdited}
	case "enter":
		return Event{EvNone}
	}
	if len([]rune(key)) != 1 {
		return Event{EvNone}
	}
	line := s.line()
	col := min(s.Cursor.Col, len(line))
	s.setLine(line[:col] + key + line[col:])
	s.Cursor.Col = col + 1
	return Event{EvEdited}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/engine/ -v`
Expected: PASS (motions + insert)

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): insert mode (i, a, typing, backspace, esc)"
```

---

### Task 4: Engine — delete and change operators (x, dw, dd, cw, cc)

**Files:**
- Modify: `internal/engine/edit.go` (real `deleteChar`, `applyOperator`; delete those stubs from `engine.go`)
- Test: `internal/engine/edit_test.go`

- [ ] **Step 1: Write the failing test (append to `edit_test.go`)**

```go
func TestDeleteAndChange(t *testing.T) {
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantBuffer []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"x deletes char", []string{"sslay"}, Pos{0, 0},
			[]string{"x"}, []string{"slay"}, Pos{0, 0}, ModeNormal},
		{"x at line end clamps cursor", []string{"ab"}, Pos{0, 1},
			[]string{"x"}, []string{"a"}, Pos{0, 0}, ModeNormal},
		{"4x deletes four chars", []string{"xxxxcore"}, Pos{0, 0},
			[]string{"4", "x"}, []string{"core"}, Pos{0, 0}, ModeNormal},
		{"dw deletes word and trailing space", []string{"the cursed goblin"}, Pos{0, 4},
			[]string{"d", "w"}, []string{"the goblin"}, Pos{0, 4}, ModeNormal},
		{"2dw deletes two words", []string{"bones and dust and shadow"}, Pos{0, 10},
			[]string{"2", "d", "w"}, []string{"bones and shadow"}, Pos{0, 10}, ModeNormal},
		{"dd deletes line", []string{"keep", "kill", "keep2"}, Pos{1, 0},
			[]string{"d", "d"}, []string{"keep", "keep2"}, Pos{1, 0}, ModeNormal},
		{"3dd deletes three lines", []string{"a", "b", "c", "d"}, Pos{0, 0},
			[]string{"3", "d", "d"}, []string{"d"}, Pos{0, 0}, ModeNormal},
		{"dd on last remaining line leaves empty buffer line", []string{"only"}, Pos{0, 0},
			[]string{"d", "d"}, []string{""}, Pos{0, 0}, ModeNormal},
		{"cw changes word and enters insert", []string{"the wolf howls"}, Pos{0, 4},
			[]string{"c", "w", "m", "o", "o", "n"}, []string{"the moon howls"}, Pos{0, 8}, ModeInsert},
		{"cc clears line and enters insert", []string{"scrap this"}, Pos{0, 3},
			[]string{"c", "c", "n", "e", "w"}, []string{"new"}, Pos{0, 3}, ModeInsert},
		{"d then invalid motion cancels", []string{"abc def"}, Pos{0, 0},
			[]string{"d", "x"}, []string{"abc def"}, Pos{0, 0}, ModeNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if !slices.Equal(s.Buffer, tt.wantBuffer) {
				t.Errorf("buffer = %q, want %q", s.Buffer, tt.wantBuffer)
			}
			if s.Cursor != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.wantCursor)
			}
			if s.Mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", s.Mode, tt.wantMode)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run TestDeleteAndChange -v`
Expected: FAIL (stubs do nothing)

- [ ] **Step 3: Implement in `edit.go`; delete the `applyOperator` and `deleteChar` stubs from `engine.go`**

```go
func (s *Simulator) deleteChar() {
	line := s.line()
	if len(line) == 0 || s.Cursor.Col >= len(line) {
		return
	}
	s.setLine(line[:s.Cursor.Col] + line[s.Cursor.Col+1:])
	s.clampCol()
}

func (s *Simulator) applyOperator(key string) Event {
	op := s.pendingOp
	s.pendingOp = ""
	n := s.opCount * s.takeCount()
	s.opCount = 0
	switch {
	case op == "d" && key == "d":
		s.snapshot()
		s.deleteLines(n)
		return Event{EvEdited}
	case op == "d" && key == "w":
		s.snapshot()
		for i := 0; i < n; i++ {
			s.deleteWord()
		}
		return Event{EvEdited}
	case op == "c" && key == "c":
		s.snapshot()
		row := s.Cursor.Row
		end := min(row+n, len(s.Buffer))
		rest := append([]string{""}, s.Buffer[end:]...)
		s.Buffer = append(s.Buffer[:row], rest...)
		s.Cursor = Pos{row, 0}
		s.Mode = ModeInsert
		return Event{EvModeChanged}
	case op == "c" && key == "w":
		s.snapshot()
		for i := 0; i < n; i++ {
			s.changeWord()
		}
		s.Mode = ModeInsert
		return Event{EvModeChanged}
	case op == "y" && key == "y":
		s.yankLines(n)
		return Event{EvNone}
	}
	return Event{EvInvalid}
}

// deleteWord removes from the cursor through the end of the word plus
// trailing spaces (dw semantics).
func (s *Simulator) deleteWord() {
	line := s.line()
	if len(line) == 0 {
		return
	}
	col := min(s.Cursor.Col, len(line))
	end := col
	for end < len(line) && line[end] != ' ' {
		end++
	}
	for end < len(line) && line[end] == ' ' {
		end++
	}
	s.setLine(line[:col] + line[end:])
	s.clampCol()
}

// changeWord removes to the end of the word only (cw semantics).
func (s *Simulator) changeWord() {
	line := s.line()
	if len(line) == 0 {
		return
	}
	col := min(s.Cursor.Col, len(line))
	end := col
	for end < len(line) && line[end] != ' ' {
		end++
	}
	s.setLine(line[:col] + line[end:])
}

func (s *Simulator) deleteLines(n int) {
	row := s.Cursor.Row
	end := min(row+n, len(s.Buffer))
	s.Buffer = append(s.Buffer[:row], s.Buffer[end:]...)
	if len(s.Buffer) == 0 {
		s.Buffer = []string{""}
	}
	if s.Cursor.Row >= len(s.Buffer) {
		s.Cursor.Row = len(s.Buffer) - 1
	}
	s.Cursor.Col = 0
}

func (s *Simulator) yankLines(n int) {
	row := s.Cursor.Row
	end := min(row+n, len(s.Buffer))
	s.yank = append([]string(nil), s.Buffer[row:end]...)
}
```

Note: `cw` leaves the cursor at the change position; typing `moon` advances it to col 8, matching the test.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/engine/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): x, dw, dd, cw, cc operators with counts"
```

---

### Task 5: Engine — yank/paste and undo

**Files:**
- Modify: `internal/engine/edit.go` (real `snapshot`, `applyUndo`, `paste`; delete those stubs from `engine.go`)
- Test: `internal/engine/edit_test.go`

- [ ] **Step 1: Write the failing test (append to `edit_test.go`)**

```go
func TestYankPasteUndo(t *testing.T) {
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantBuffer []string
		wantCursor Pos
	}{
		{"yy p duplicates line below", []string{"echo"}, Pos{0, 0},
			[]string{"y", "y", "p"}, []string{"echo", "echo"}, Pos{1, 0}},
		{"yy j p pastes after other line", []string{"chant", "deep"}, Pos{0, 0},
			[]string{"y", "y", "j", "p"}, []string{"chant", "deep", "chant"}, Pos{2, 0}},
		{"2yy p pastes two lines", []string{"a", "b", "c"}, Pos{0, 0},
			[]string{"2", "y", "y", "p"}, []string{"a", "a", "b", "b", "c"}, Pos{1, 0}},
		{"p with empty register is a no-op", []string{"a"}, Pos{0, 0},
			[]string{"p"}, []string{"a"}, Pos{0, 0}},
		{"u undoes x", []string{"abc"}, Pos{0, 0},
			[]string{"x", "u"}, []string{"abc"}, Pos{0, 0}},
		{"u undoes dd", []string{"a", "b"}, Pos{0, 0},
			[]string{"d", "d", "u"}, []string{"a", "b"}, Pos{0, 0}},
		{"u undoes whole insert session", []string{"ab"}, Pos{0, 0},
			[]string{"i", "X", "Y", "esc", "u"}, []string{"ab"}, Pos{0, 0}},
		{"u with nothing to undo is a no-op", []string{"ab"}, Pos{0, 1},
			[]string{"u"}, []string{"ab"}, Pos{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if !slices.Equal(s.Buffer, tt.wantBuffer) {
				t.Errorf("buffer = %q, want %q", s.Buffer, tt.wantBuffer)
			}
			if s.Cursor != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.wantCursor)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run TestYankPasteUndo -v`
Expected: FAIL (paste/undo stubs do nothing)

- [ ] **Step 3: Implement in `edit.go`; delete `snapshot`, `applyUndo`, `paste` stubs from `engine.go`**

```go
func (s *Simulator) snapshot() {
	buf := make([]string, len(s.Buffer))
	copy(buf, s.Buffer)
	s.undo = append(s.undo, snapshot{buffer: buf, cursor: s.Cursor})
}

func (s *Simulator) applyUndo() Event {
	if len(s.undo) == 0 {
		return Event{EvNone}
	}
	last := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	s.Buffer = last.buffer
	s.Cursor = last.cursor
	return Event{EvEdited}
}

func (s *Simulator) paste() Event {
	if len(s.yank) == 0 {
		return Event{EvNone}
	}
	s.snapshot()
	row := s.Cursor.Row
	out := make([]string, 0, len(s.Buffer)+len(s.yank))
	out = append(out, s.Buffer[:row+1]...)
	out = append(out, s.yank...)
	out = append(out, s.Buffer[row+1:]...)
	s.Buffer = out
	s.Cursor = Pos{row + 1, 0}
	return Event{EvEdited}
}
```

(`i`/`a` already snapshot before entering insert mode — Task 2 — so one `u` undoes the whole insert session.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/engine/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): yank/paste and snapshot-based undo"
```

---

### Task 6: Engine — search mode and n

**Files:**
- Create: `internal/engine/search.go` (real `pressSearch`, `jumpToMatch`; delete those stubs from `engine.go`)
- Test: `internal/engine/search_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import "testing"

func TestSearch(t *testing.T) {
	buf := []string{"trace node1 active", "trace node2 active", "trace node3 active"}
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"/term enter jumps to first match", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter"}, Pos{0, 6}, ModeNormal},
		{"n repeats forward", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter", "n"}, Pos{1, 6}, ModeNormal},
		{"n wraps around", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter", "n", "n", "n"}, Pos{0, 6}, ModeNormal},
		{"esc cancels search", buf, Pos{0, 2},
			[]string{"/", "z", "esc"}, Pos{0, 2}, ModeNormal},
		{"backspace edits query", buf, Pos{0, 0},
			[]string{"/", "n", "z", "backspace", "o", "d", "e", "enter"}, Pos{0, 6}, ModeNormal},
		{"no match stays put", buf, Pos{0, 2},
			[]string{"/", "q", "q", "enter"}, Pos{0, 2}, ModeNormal},
		{"n without prior search is no-op", buf, Pos{0, 2},
			[]string{"n"}, Pos{0, 2}, ModeNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if s.Cursor != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.wantCursor)
			}
			if s.Mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", s.Mode, tt.wantMode)
			}
		})
	}
}

func TestSearchQueryVisible(t *testing.T) {
	s := New([]string{"abc"}, Pos{0, 0})
	s.Press("/")
	s.Press("a")
	s.Press("b")
	if s.SearchQuery != "ab" {
		t.Errorf("SearchQuery = %q, want %q", s.SearchQuery, "ab")
	}
	if s.Mode != ModeSearch {
		t.Errorf("mode = %v, want ModeSearch", s.Mode)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run TestSearch -v`
Expected: FAIL

- [ ] **Step 3: Implement — create `internal/engine/search.go`; delete `pressSearch` and `jumpToMatch` stubs from `engine.go`**

```go
package engine

import "strings"

func (s *Simulator) pressSearch(key string) Event {
	switch key {
	case "esc":
		s.Mode = ModeNormal
		s.SearchQuery = ""
		return Event{EvModeChanged}
	case "enter":
		s.lastSearch = s.SearchQuery
		s.SearchQuery = ""
		s.Mode = ModeNormal
		if s.jumpToMatch(s.lastSearch, true) {
			return Event{EvSearchJumped}
		}
		return Event{EvNone}
	case "backspace":
		if len(s.SearchQuery) > 0 {
			s.SearchQuery = s.SearchQuery[:len(s.SearchQuery)-1]
		}
		return Event{EvNone}
	}
	if len([]rune(key)) == 1 {
		s.SearchQuery += key
	}
	return Event{EvNone}
}

// jumpToMatch finds term forward from the cursor (exclusive when fromNext),
// wrapping around the buffer. Returns true and moves the cursor on a hit.
func (s *Simulator) jumpToMatch(term string, fromNext bool) bool {
	if term == "" {
		return false
	}
	rows := len(s.Buffer)
	startCol := s.Cursor.Col
	if fromNext {
		startCol++
	}
	for i := 0; i <= rows; i++ {
		row := (s.Cursor.Row + i) % rows
		line := s.Buffer[row]
		col := 0
		if i == 0 {
			col = min(startCol, len(line))
		}
		if idx := strings.Index(line[col:], term); idx >= 0 {
			s.Cursor = Pos{row, col + idx}
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/engine/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): incremental search with / and n"
```

---

### Task 7: Engine — allowed keys and invalid events

**Files:**
- Modify: `internal/engine/engine.go` (already implemented in Task 2 — this task locks behavior in with tests)
- Test: `internal/engine/engine_test.go`

- [ ] **Step 1: Write the failing test (append to `engine_test.go`)**

```go
func TestAllowedKeysAndInvalid(t *testing.T) {
	allow := func(keys ...string) map[string]bool {
		m := map[string]bool{}
		for _, k := range keys {
			m[k] = true
		}
		return m
	}
	t.Run("disallowed normal-mode key is invalid", func(t *testing.T) {
		s := New([]string{"abc"}, Pos{0, 0})
		s.AllowedKeys = allow("h", "l")
		if ev := s.Press("x"); ev.Kind != EvInvalid {
			t.Errorf("Press(x) = %v, want EvInvalid", ev.Kind)
		}
		if s.Buffer[0] != "abc" {
			t.Errorf("buffer mutated by disallowed key: %q", s.Buffer[0])
		}
	})
	t.Run("esc is never penalized", func(t *testing.T) {
		s := New([]string{"abc"}, Pos{0, 0})
		s.AllowedKeys = allow("h")
		if ev := s.Press("esc"); ev.Kind == EvInvalid {
			t.Error("esc must not be invalid")
		}
	})
	t.Run("insert typing ignores AllowedKeys", func(t *testing.T) {
		s := New([]string{""}, Pos{0, 0})
		s.AllowedKeys = allow("i")
		s.Press("i")
		if ev := s.Press("z"); ev.Kind != EvEdited {
			t.Errorf("insert typing = %v, want EvEdited", ev.Kind)
		}
	})
	t.Run("nil AllowedKeys permits everything", func(t *testing.T) {
		s := New([]string{"abc"}, Pos{0, 0})
		if ev := s.Press("x"); ev.Kind != EvEdited {
			t.Errorf("Press(x) = %v, want EvEdited", ev.Kind)
		}
	})
	t.Run("unknown normal key is invalid", func(t *testing.T) {
		s := New([]string{"abc"}, Pos{0, 0})
		if ev := s.Press("Z"); ev.Kind != EvInvalid {
			t.Errorf("Press(Z) = %v, want EvInvalid", ev.Kind)
		}
	})
	t.Run("disallowed key clears pending operator", func(t *testing.T) {
		s := New([]string{"abc def"}, Pos{0, 0})
		s.AllowedKeys = allow("d", "w")
		s.Press("d")
		s.Press("q") // invalid
		if got := s.Pending(); got != "" {
			t.Errorf("pending = %q, want empty after invalid key", got)
		}
	})
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./internal/engine/ -run TestAllowedKeysAndInvalid -v`
Expected: PASS already (behavior shipped in Task 2). If any subtest fails, fix `pressNormal` until green — the contract in this test is authoritative.

- [ ] **Step 3: Commit**

```bash
git add internal/engine
git commit -m "test(engine): lock in allowed-key and invalid-key contract"
```

---

### Task 8: Engine — goal validators

**Files:**
- Create: `internal/engine/goal.go`
- Test: `internal/engine/goal_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import "testing"

func TestGoals(t *testing.T) {
	tests := []struct {
		name  string
		goal  Goal
		setup func() *Simulator
		want  bool
	}{
		{"cursorOnWord hit", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 6}) }, true},
		{"cursorOnWord middle of word counts", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 8}) }, true},
		{"cursorOnWord miss", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 0}) }, false},
		{"bufferEquals hit", Goal{Type: "bufferEquals", Lines: []string{"a", "b"}},
			func() *Simulator { return New([]string{"a", "b"}, Pos{0, 0}) }, true},
		{"bufferEquals miss", Goal{Type: "bufferEquals", Lines: []string{"a"}},
			func() *Simulator { return New([]string{"a", "b"}, Pos{0, 0}) }, false},
		{"lineDeleted hit", Goal{Type: "lineDeleted", Line: "gone"},
			func() *Simulator { return New([]string{"keep"}, Pos{0, 0}) }, true},
		{"lineDeleted miss", Goal{Type: "lineDeleted", Line: "gone"},
			func() *Simulator { return New([]string{"keep", "gone"}, Pos{0, 0}) }, false},
		{"wordDeleted hit", Goal{Type: "wordDeleted", Word: "cursed"},
			func() *Simulator { return New([]string{"the goblin"}, Pos{0, 0}) }, true},
		{"wordDeleted miss", Goal{Type: "wordDeleted", Word: "cursed"},
			func() *Simulator { return New([]string{"the cursed goblin"}, Pos{0, 0}) }, false},
		{"containsText hit", Goal{Type: "containsText", Text: "I am ready"},
			func() *Simulator { return New([]string{"I am ready now"}, Pos{0, 0}) }, true},
		{"containsText miss", Goal{Type: "containsText", Text: "I am ready"},
			func() *Simulator { return New([]string{"am ready"}, Pos{0, 0}) }, false},
		{"searchMatchActive hit", Goal{Type: "searchMatchActive", Term: "backdoor"},
			func() *Simulator {
				s := New([]string{"a backdoor hides"}, Pos{0, 0})
				for _, k := range []string{"/", "b", "a", "c", "k", "d", "o", "o", "r", "enter"} {
					s.Press(k)
				}
				return s
			}, true},
		{"searchMatchActive miss without search", Goal{Type: "searchMatchActive", Term: "backdoor"},
			func() *Simulator { return New([]string{"a backdoor hides"}, Pos{0, 2}) }, false},
		{"unknown goal type is never met", Goal{Type: "bogus"},
			func() *Simulator { return New([]string{"a"}, Pos{0, 0}) }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.goal.Met(tt.setup()); got != tt.want {
				t.Errorf("Met() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/ -run TestGoals -v`
Expected: FAIL (Goal undefined)

- [ ] **Step 3: Implement — `internal/engine/goal.go`**

```go
package engine

import (
	"slices"
	"strings"
)

// Goal is a challenge success condition, decoded straight from lesson JSON.
type Goal struct {
	Type  string   `json:"type"`
	Word  string   `json:"word,omitempty"`
	Line  string   `json:"line,omitempty"`
	Lines []string `json:"lines,omitempty"`
	Text  string   `json:"text,omitempty"`
	Term  string   `json:"term,omitempty"`
}

// GoalTypes lists every valid Goal.Type, for content validation.
var GoalTypes = []string{
	"cursorOnWord", "bufferEquals", "lineDeleted",
	"wordDeleted", "containsText", "searchMatchActive",
}

func (g Goal) Met(s *Simulator) bool {
	switch g.Type {
	case "cursorOnWord":
		return s.WordAtCursor() == g.Word
	case "bufferEquals":
		return slices.Equal(s.Buffer, g.Lines)
	case "lineDeleted":
		return !slices.Contains(s.Buffer, g.Line)
	case "wordDeleted":
		for _, l := range s.Buffer {
			if strings.Contains(l, g.Word) {
				return false
			}
		}
		return true
	case "containsText":
		for _, l := range s.Buffer {
			if strings.Contains(l, g.Text) {
				return true
			}
		}
		return false
	case "searchMatchActive":
		line := s.line()
		if s.Cursor.Col >= len(line) {
			return false
		}
		return s.lastSearch == g.Term && strings.HasPrefix(line[s.Cursor.Col:], g.Term)
	}
	return false
}

// WordAtCursor returns the space-delimited word under the cursor, or "".
func (s *Simulator) WordAtCursor() string {
	line := s.line()
	col := s.Cursor.Col
	if col >= len(line) || line[col] == ' ' {
		return ""
	}
	start := col
	for start > 0 && line[start-1] != ' ' {
		start--
	}
	end := col
	for end < len(line) && line[end] != ' ' {
		end++
	}
	return line[start:end]
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/engine/ -v`
Expected: PASS (entire engine suite)

- [ ] **Step 5: Commit**

```bash
git add internal/engine
git commit -m "feat(engine): six goal validators"
```

---

### Task 9: Content — models, loader, integrity test, Act I lessons

**Files:**
- Create: `internal/content/model.go`, `internal/content/loader.go`
- Create: `assets/lessons/act1-01-the-two-stances.json` … `act1-04-the-great-leaps.json`
- Delete: `assets/lessons/placeholder.json`
- Test: `internal/content/loader_test.go`

- [ ] **Step 1: Write the failing test**

```go
package content

import "testing"

func TestAllLoadsAndValidates(t *testing.T) {
	lessons, err := All()
	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	if len(lessons) == 0 {
		t.Fatal("no lessons loaded")
	}
	seen := map[string]bool{}
	for i, l := range lessons {
		if l.ID == "" || l.Title == "" {
			t.Errorf("lesson %d missing id/title", i)
		}
		if l.Act < 1 || l.Act > 3 {
			t.Errorf("%s: act %d out of range", l.ID, l.Act)
		}
		if i > 0 {
			prev := lessons[i-1]
			if l.Act < prev.Act || (l.Act == prev.Act && l.Order <= prev.Order) {
				t.Errorf("%s not sorted after %s", l.ID, prev.ID)
			}
		}
		if len(l.Challenges) == 0 {
			t.Errorf("%s has no challenges", l.ID)
		}
		for _, ch := range append([]Challenge{}, l.Challenges...) {
			if seen[ch.ID] {
				t.Errorf("duplicate challenge id %s", ch.ID)
			}
			seen[ch.ID] = true
			validateChallenge(t, l.ID, ch, true)
		}
		if l.Boss != nil {
			if l.Boss.TimeLimitSec < 30 {
				t.Errorf("%s boss time limit too low", l.ID)
			}
			if l.Boss.XP <= 0 {
				t.Errorf("%s boss has no xp", l.ID)
			}
			if len(l.Boss.Steps) == 0 {
				t.Errorf("%s boss has no steps", l.ID)
			}
			for _, st := range l.Boss.Steps {
				validateChallenge(t, l.ID+":boss", st, false)
			}
		}
	}
}

func validateChallenge(t *testing.T, owner string, ch Challenge, needsParXP bool) {
	t.Helper()
	if len(ch.Buffer) == 0 {
		t.Errorf("%s/%s: empty buffer", owner, ch.ID)
	}
	row, col := ch.Cursor[0], ch.Cursor[1]
	if row < 0 || row >= len(ch.Buffer) || col < 0 || (len(ch.Buffer) > row && col > len(ch.Buffer[row])) {
		t.Errorf("%s/%s: cursor %v out of bounds", owner, ch.ID, ch.Cursor)
	}
	valid := false
	for _, gt := range goalTypes() {
		if ch.Goal.Type == gt {
			valid = true
		}
	}
	if !valid {
		t.Errorf("%s/%s: unknown goal type %q", owner, ch.ID, ch.Goal.Type)
	}
	if needsParXP {
		if ch.Par < 1 {
			t.Errorf("%s/%s: par must be >= 1", owner, ch.ID)
		}
		if ch.XP <= 0 {
			t.Errorf("%s/%s: xp must be > 0", owner, ch.ID)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/content/ -v`
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Implement — `internal/content/model.go`**

```go
// Package content loads the embedded lesson JSON files.
package content

import "nvim-quest/internal/engine"

type Lesson struct {
	ID         string      `json:"id"`
	Act        int         `json:"act"`
	Order      int         `json:"order"`
	Title      string      `json:"title"`
	Story      string      `json:"story"`
	Challenges []Challenge `json:"challenges"`
	Boss       *Boss       `json:"boss,omitempty"`
}

type Challenge struct {
	ID          string      `json:"id"`
	Intro       string      `json:"intro"`
	Buffer      []string    `json:"buffer"`
	Cursor      [2]int      `json:"cursor"`
	Goal        engine.Goal `json:"goal"`
	Par         int         `json:"par"`
	XP          int         `json:"xp"`
	Hint        string      `json:"hint"`
	NewKeys     []string    `json:"newKeys,omitempty"`
	AllowedKeys []string    `json:"allowedKeys,omitempty"`
}

type Boss struct {
	Name         string      `json:"name"`
	Taunt        string      `json:"taunt"`
	TimeLimitSec int         `json:"timeLimitSec"`
	XP           int         `json:"xp"`
	Steps        []Challenge `json:"steps"`
}

func goalTypes() []string { return engine.GoalTypes }
```

- [ ] **Step 4: Implement — `internal/content/loader.go`**

```go
package content

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"nvim-quest/assets"
)

// All returns every embedded lesson, sorted by (act, order).
func All() ([]Lesson, error) {
	entries, err := assets.Lessons.ReadDir("lessons")
	if err != nil {
		return nil, fmt.Errorf("read embedded lessons: %w", err)
	}
	var lessons []Lesson
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := assets.Lessons.ReadFile("lessons/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		var l Lesson
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		lessons = append(lessons, l)
	}
	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].Act != lessons[j].Act {
			return lessons[i].Act < lessons[j].Act
		}
		return lessons[i].Order < lessons[j].Order
	})
	return lessons, nil
}
```

- [ ] **Step 5: Write the Act I lesson files (and `rm assets/lessons/placeholder.json`)**

`assets/lessons/act1-01-the-two-stances.json`:

```json
{
  "id": "act1-01-the-two-stances",
  "act": 1, "order": 1,
  "title": "The Two Stances",
  "story": "Sensei: \"The blade cursor knows two stances. Normal, where you command. Insert, where you write. Master the switch between them.\"",
  "challenges": [
    {
      "id": "a1l1c1",
      "intro": "Declare yourself ready. Enter the insert stance and add 'I ' at the front.",
      "buffer": ["am ready"], "cursor": [0, 0],
      "goal": { "type": "containsText", "text": "I am ready" },
      "par": 3, "xp": 50,
      "hint": "Press i to enter insert stance, type the text, then Esc to return.",
      "newKeys": ["i", "esc"],
      "allowedKeys": ["i"]
    },
    {
      "id": "a1l1c2",
      "intro": "A word is missing between the breaths. Insert 'in' at the cursor.",
      "buffer": ["breathe  out"], "cursor": [0, 8],
      "goal": { "type": "bufferEquals", "lines": ["breathe in out"] },
      "par": 3, "xp": 50,
      "hint": "i inserts before the cursor.",
      "allowedKeys": ["i"]
    }
  ]
}
```

`assets/lessons/act1-02-first-steps.json`:

```json
{
  "id": "act1-02-first-steps",
  "act": 1, "order": 2,
  "title": "First Steps",
  "story": "Sensei: \"Every journey begins with a single step. h left, j down, k up, l right. Walk.\"",
  "challenges": [
    {
      "id": "a1l2c1",
      "intro": "Walk to the master.",
      "buffer": ["walk the path", "to the master"], "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "master" },
      "par": 8, "xp": 50,
      "hint": "j moves down a line, l moves right one step.",
      "newKeys": ["h", "j", "k", "l"],
      "allowedKeys": ["h", "j", "k", "l"]
    },
    {
      "id": "a1l2c2",
      "intro": "Climb back up to the summit.",
      "buffer": ["the summit waits", "climb back up", "from the valley"], "cursor": [2, 0],
      "goal": { "type": "cursorOnWord", "word": "summit" },
      "par": 6, "xp": 50,
      "hint": "k climbs up, l steps right.",
      "allowedKeys": ["h", "j", "k", "l"]
    }
  ]
}
```

`assets/lessons/act1-03-way-of-the-word.json`:

```json
{
  "id": "act1-03-way-of-the-word",
  "act": 1, "order": 3,
  "title": "Way of the Word",
  "story": "Sensei: \"The wise cursor moves by words, not steps. w leaps forward, b leaps back.\"",
  "challenges": [
    {
      "id": "a1l3c1",
      "intro": "Reach the temple in as few strides as you can.",
      "buffer": ["the path to the temple is long"], "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "temple" },
      "par": 4, "xp": 50,
      "hint": "w jumps to the start of the next word.",
      "newKeys": ["w", "b"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b"]
    },
    {
      "id": "a1l3c2",
      "intro": "Wisdom lies behind you. Leap back to it.",
      "buffer": ["wisdom flows backward"], "cursor": [0, 20],
      "goal": { "type": "cursorOnWord", "word": "wisdom" },
      "par": 3, "xp": 50,
      "hint": "b jumps to the start of the previous word.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b"]
    }
  ]
}
```

`assets/lessons/act1-04-the-great-leaps.json`:

```json
{
  "id": "act1-04-the-great-leaps",
  "act": 1, "order": 4,
  "title": "The Great Leaps",
  "story": "Sensei: \"Why walk when you can fly? 0 to the line's dawn, $ to its dusk. gg to the summit, G to the roots.\"",
  "challenges": [
    {
      "id": "a1l4c1",
      "intro": "Dive to the river's mouth at the very bottom.",
      "buffer": ["mountain peak above", "mist in the middle", "river mouth below"], "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "mouth" },
      "par": 2, "xp": 50,
      "hint": "G drops to the last line; w hops to the next word.",
      "newKeys": ["0", "$", "gg", "G"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G"]
    },
    {
      "id": "a1l4c2",
      "intro": "Soar back to the peak, to the word at the line's end.",
      "buffer": ["mountain peak above", "mist in the middle", "river mouth below"], "cursor": [2, 8],
      "goal": { "type": "cursorOnWord", "word": "above" },
      "par": 3, "xp": 50,
      "hint": "gg flies to the first line, $ to the line's end.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G"]
    }
  ],
  "boss": {
    "name": "Sensei's Trial",
    "taunt": "Show me the way of motion, student. The hourglass empties.",
    "timeLimitSec": 60,
    "xp": 200,
    "steps": [
      {
        "id": "a1boss1",
        "intro": "Seek the stone.",
        "buffer": ["seek the stone within the garden"], "cursor": [0, 0],
        "goal": { "type": "cursorOnWord", "word": "stone" }
      },
      {
        "id": "a1boss2",
        "intro": "Descend to the deepest roots.",
        "buffer": ["high above clouds", "between two worlds", "deep below roots"], "cursor": [0, 0],
        "goal": { "type": "cursorOnWord", "word": "roots" }
      },
      {
        "id": "a1boss3",
        "intro": "Return to the beginning.",
        "buffer": ["return to the beginning now"], "cursor": [0, 25],
        "goal": { "type": "cursorOnWord", "word": "return" }
      }
    ]
  }
}
```

```bash
rm assets/lessons/placeholder.json
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/content/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/content assets
git commit -m "feat(content): models, embedded loader, Act I lessons + boss"
```

---

### Task 10: Content — Act II and Act III lessons

**Files:**
- Create: `assets/lessons/act2-05-hall-of-insertion.json` … `act3-10-power-surge.json`
- Modify: `internal/content/loader_test.go` (assert final counts)

- [ ] **Step 1: Extend the test (append inside `TestAllLoadsAndValidates`, after the loop)**

```go
	if len(lessons) != 10 {
		t.Errorf("expected 10 lessons, got %d", len(lessons))
	}
	bosses := 0
	lastInAct := map[int]string{}
	for _, l := range lessons {
		lastInAct[l.Act] = l.ID
		if l.Boss != nil {
			bosses++
		}
	}
	if bosses != 3 {
		t.Errorf("expected 3 bosses, got %d", bosses)
	}
	for act, id := range lastInAct {
		for _, l := range lessons {
			if l.ID == id && l.Boss == nil {
				t.Errorf("act %d final lesson %s must carry the boss", act, id)
			}
		}
	}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/content/ -v`
Expected: FAIL ("expected 10 lessons, got 4")

- [ ] **Step 3: Write the Act II lesson files**

`assets/lessons/act2-05-hall-of-insertion.json`:

```json
{
  "id": "act2-05-hall-of-insertion",
  "act": 2, "order": 5,
  "title": "Hall of Insertion",
  "story": "Torchlight flickers. Carvings on the crypt wall are unfinished — and the door will not open until the words are whole.",
  "challenges": [
    {
      "id": "a2l5c1",
      "intro": "Finish the carving: the crypt door is... open. Append the final word.",
      "buffer": ["the crypt door is "], "cursor": [0, 17],
      "goal": { "type": "containsText", "text": "the crypt door is open" },
      "par": 5, "xp": 60,
      "hint": "a appends after the cursor — perfect for the end of a line.",
      "newKeys": ["a"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u"]
    },
    {
      "id": "a2l5c2",
      "intro": "An old word crumbled from the middle. Restore 'old' before the torch.",
      "buffer": ["light the  torch"], "cursor": [0, 10],
      "goal": { "type": "bufferEquals", "lines": ["light the old torch"] },
      "par": 4, "xp": 60,
      "hint": "i inserts before the cursor.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u"]
    }
  ]
}
```

`assets/lessons/act2-06-the-deletion-pits.json`:

```json
{
  "id": "act2-06-the-deletion-pits",
  "act": 2, "order": 6,
  "title": "The Deletion Pits",
  "story": "Cursed glyphs crawl across the walls. Your blade can cut a letter (x), a word (dw), or an entire line (dd).",
  "challenges": [
    {
      "id": "a2l6c1",
      "intro": "A stray letter corrupts the war cry. Cut it.",
      "buffer": ["sslay the beast"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["slay the beast"] },
      "par": 1, "xp": 60,
      "hint": "x deletes the character under the cursor.",
      "newKeys": ["x", "dw", "dd"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d"]
    },
    {
      "id": "a2l6c2",
      "intro": "Lift the curse from the goblin. Delete the cursed word.",
      "buffer": ["the cursed goblin lurks"], "cursor": [0, 4],
      "goal": { "type": "bufferEquals", "lines": ["the goblin lurks"] },
      "par": 2, "xp": 60,
      "hint": "dw deletes from the cursor through the end of the word.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d"]
    },
    {
      "id": "a2l6c3",
      "intro": "One inscription must burn. Erase the entire middle line.",
      "buffer": ["keep this relic", "burn this line", "keep this too"], "cursor": [1, 0],
      "goal": { "type": "lineDeleted", "line": "burn this line" },
      "par": 2, "xp": 60,
      "hint": "dd deletes the whole current line.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d"]
    }
  ]
}
```

`assets/lessons/act2-07-echo-chamber.json`:

```json
{
  "id": "act2-07-echo-chamber",
  "act": 2, "order": 7,
  "title": "Echo Chamber",
  "story": "This chamber repeats every sound. Yank a line into memory (yy) and let it echo back (p).",
  "challenges": [
    {
      "id": "a2l7c1",
      "intro": "Make the chamber echo. Duplicate the word below itself.",
      "buffer": ["echo"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["echo", "echo"] },
      "par": 3, "xp": 60,
      "hint": "yy copies the line, p pastes it below.",
      "newKeys": ["yy", "p"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p"]
    },
    {
      "id": "a2l7c2",
      "intro": "Carry the chant down past the deep and let it echo at the bottom.",
      "buffer": ["the chant repeats", "in the deep"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["the chant repeats", "in the deep", "the chant repeats"] },
      "par": 4, "xp": 60,
      "hint": "Yank with yy, move down with j, paste with p.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p"]
    }
  ]
}
```

`assets/lessons/act2-08-the-shapeshifter.json`:

```json
{
  "id": "act2-08-the-shapeshifter",
  "act": 2, "order": 8,
  "title": "The Shapeshifter",
  "story": "A shapeshifter haunts these halls. Change a word in one motion (cw) or remake a whole line (cc).",
  "challenges": [
    {
      "id": "a2l8c1",
      "intro": "The wolf is an illusion. Change it into the moon.",
      "buffer": ["the wolf howls"], "cursor": [0, 4],
      "goal": { "type": "bufferEquals", "lines": ["the moon howls"] },
      "par": 6, "xp": 60,
      "hint": "cw deletes the word and drops you straight into insert stance.",
      "newKeys": ["cw", "cc"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c"]
    },
    {
      "id": "a2l8c2",
      "intro": "This line is beyond saving. Remake it as a single word: reborn",
      "buffer": ["scrap this entire line"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["reborn"] },
      "par": 8, "xp": 60,
      "hint": "cc clears the whole line and enters insert stance.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c"]
    }
  ],
  "boss": {
    "name": "The Gravewright",
    "taunt": "Your edits cannot save you, little scribe. The tomb closes.",
    "timeLimitSec": 90,
    "xp": 250,
    "steps": [
      {
        "id": "a2boss1",
        "intro": "Cut the doubled curse letter.",
        "buffer": ["the ccurse holds"], "cursor": [0, 4],
        "goal": { "type": "bufferEquals", "lines": ["the curse holds"] }
      },
      {
        "id": "a2boss2",
        "intro": "Sweep the dust away — two words must go.",
        "buffer": ["bones and dust and shadow"], "cursor": [0, 10],
        "goal": { "type": "bufferEquals", "lines": ["bones and shadow"] }
      },
      {
        "id": "a2boss3",
        "intro": "The Gravewright rises... change his fate to 'falls'.",
        "buffer": ["the gravewright rises"], "cursor": [0, 16],
        "goal": { "type": "bufferEquals", "lines": ["the gravewright falls"] }
      },
      {
        "id": "a2boss4",
        "intro": "Seal the tomb with a doubled binding.",
        "buffer": ["seal the tomb"], "cursor": [0, 0],
        "goal": { "type": "bufferEquals", "lines": ["seal the tomb", "seal the tomb"] }
      }
    ]
  }
}
```

- [ ] **Step 4: Write the Act III lesson files**

`assets/lessons/act3-09-trace-evasion.json`:

```json
{
  "id": "act3-09-trace-evasion",
  "act": 3, "order": 9,
  "title": "Trace Evasion",
  "story": "You are inside the Grid. Eyes everywhere. Search (/) finds anything instantly — n hops to the next hit before the trace locks on.",
  "challenges": [
    {
      "id": "a3l9c1",
      "intro": "Locate the backdoor before the trace completes.",
      "buffer": ["the grid hums with data", "a backdoor hides in the stream"], "cursor": [0, 0],
      "goal": { "type": "searchMatchActive", "term": "backdoor" },
      "par": 10, "xp": 70,
      "hint": "Type /backdoor then Enter.",
      "newKeys": ["/", "n"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c", "/", "n"]
    },
    {
      "id": "a3l9c2",
      "intro": "Three nodes share a signature. Hop matches until you reach node3.",
      "buffer": ["trace node1 active", "trace node2 active", "trace node3 active"], "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "node3" },
      "par": 8, "xp": 70,
      "hint": "Search /node then press n to jump to the next match.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c", "/", "n"]
    }
  ]
}
```

`assets/lessons/act3-10-power-surge.json`:

```json
{
  "id": "act3-10-power-surge",
  "act": 3, "order": 10,
  "title": "Power Surge",
  "story": "Why press a key four times when a number presses it for you? Counts multiply any motion or edit: 4w, 3dd, 4x.",
  "challenges": [
    {
      "id": "a3l10c1",
      "intro": "Surge four words forward in a single command.",
      "buffer": ["one two three four five six"], "cursor": [0, 0],
      "goal": { "type": "cursorOnWord", "word": "five" },
      "par": 2, "xp": 70,
      "hint": "4w = w four times.",
      "newKeys": ["counts"],
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c", "/", "n", "2", "3", "4"]
    },
    {
      "id": "a3l10c2",
      "intro": "Three decoy lines guard the core. Drop them all at once.",
      "buffer": ["zap", "zap", "zap", "the core stands"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["the core stands"] },
      "par": 3, "xp": 70,
      "hint": "3dd deletes three lines.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c", "/", "n", "2", "3", "4"]
    },
    {
      "id": "a3l10c3",
      "intro": "Four corrupt bytes mask the core. Purge them in one strike.",
      "buffer": ["xxxxcore"], "cursor": [0, 0],
      "goal": { "type": "bufferEquals", "lines": ["core"] },
      "par": 2, "xp": 70,
      "hint": "4x deletes four characters.",
      "allowedKeys": ["h", "j", "k", "l", "w", "b", "0", "$", "g", "G", "i", "a", "u", "x", "d", "y", "p", "c", "/", "n", "2", "3", "4"]
    }
  ],
  "boss": {
    "name": "The Grid Core",
    "taunt": "I am the Grid. You are a syntax error.",
    "timeLimitSec": 120,
    "xp": 300,
    "steps": [
      {
        "id": "a3boss1",
        "intro": "Bring down the firewall line.",
        "buffer": ["firewall firewall firewall", "the kernel sleeps"], "cursor": [0, 0],
        "goal": { "type": "lineDeleted", "line": "firewall firewall firewall" }
      },
      {
        "id": "a3boss2",
        "intro": "Rewrite your fate: access granted.",
        "buffer": ["access denied"], "cursor": [0, 7],
        "goal": { "type": "bufferEquals", "lines": ["access granted"] }
      },
      {
        "id": "a3boss3",
        "intro": "Surge past four nodes to the breach.",
        "buffer": ["node node node node breach"], "cursor": [0, 0],
        "goal": { "type": "cursorOnWord", "word": "breach" }
      },
      {
        "id": "a3boss4",
        "intro": "Find where the core burns.",
        "buffer": ["the grid core burns", "deep in silicon"], "cursor": [0, 0],
        "goal": { "type": "searchMatchActive", "term": "burns" }
      }
    ]
  }
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/content/ -v`
Expected: PASS (10 lessons, 3 bosses)

- [ ] **Step 6: Commit**

```bash
git add internal/content assets
git commit -m "feat(content): Act II and Act III lessons with bosses"
```

---

### Task 11: Game mechanics — stars, combo, XP, levels, badges

**Files:**
- Create: `internal/game/game.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Write the failing test**

```go
package game

import "testing"

func TestStars(t *testing.T) {
	tests := []struct{ keys, par, want int }{
		{3, 4, 3}, {4, 4, 3}, {5, 4, 2}, {6, 4, 2}, {7, 4, 1}, {20, 4, 1},
	}
	for _, tt := range tests {
		if got := Stars(tt.keys, tt.par); got != tt.want {
			t.Errorf("Stars(%d, %d) = %d, want %d", tt.keys, tt.par, got, tt.want)
		}
	}
}

func TestNextCombo(t *testing.T) {
	tests := []struct{ in, want int }{{1, 2}, {4, 5}, {5, 5}}
	for _, tt := range tests {
		if got := NextCombo(tt.in); got != tt.want {
			t.Errorf("NextCombo(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestXP(t *testing.T) {
	if got := XP(50, 3); got != 150 {
		t.Errorf("XP(50, 3) = %d, want 150", got)
	}
}

func TestLevelForXP(t *testing.T) {
	tests := []struct{ xp, want int }{
		{0, 1}, {99, 1}, {100, 2}, {299, 2}, {300, 3}, {599, 3}, {600, 4},
	}
	for _, tt := range tests {
		if got := LevelForXP(tt.xp); got != tt.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tt.xp, got, tt.want)
		}
	}
}

func TestBirdieEarned(t *testing.T) {
	stars := map[string]int{}
	for i := 0; i < 9; i++ {
		stars[string(rune('a'+i))] = 3
	}
	if BirdieEarned(stars) {
		t.Error("9 three-star clears must not earn Birdie")
	}
	stars["j"] = 3
	if !BirdieEarned(stars) {
		t.Error("10 three-star clears must earn Birdie")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/game/ -v`
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Implement — `internal/game/game.go`**

```go
// Package game holds the scoring and progression math.
package game

const MaxCombo = 5

// Stars rates a clear: at or under par 3, within par+2 2, otherwise 1.
func Stars(keystrokes, par int) int {
	switch {
	case keystrokes <= par:
		return 3
	case keystrokes <= par+2:
		return 2
	default:
		return 1
	}
}

// NextCombo raises the multiplier after a clean clear, capped at MaxCombo.
func NextCombo(c int) int {
	if c < MaxCombo {
		return c + 1
	}
	return MaxCombo
}

// XP is the base award scaled by the active combo multiplier.
func XP(base, combo int) int { return base * combo }

// LevelForXP: reaching level N+1 costs 100*N XP beyond level N
// (thresholds 0, 100, 300, 600, 1000, ...).
func LevelForXP(xp int) int {
	level, need := 1, 0
	for {
		need += 100 * level
		if xp < need {
			return level
		}
		level++
	}
}

// Badge names.
const (
	BadgeFirstSteps  = "First Steps"
	BadgeBirdie      = "Birdie"
	BadgeUntouchable = "Untouchable"
	BadgeGridBreaker = "Grid Breaker"
)

// ActBadge names the badge for completing an act's boss.
func ActBadge(act int) string {
	switch act {
	case 1:
		return "Dojo Graduate"
	case 2:
		return "Crypt Conqueror"
	case 3:
		return "Grid Runner"
	}
	return ""
}

// BirdieEarned reports whether 10+ challenges have been cleared at 3 stars.
func BirdieEarned(stars map[string]int) bool {
	count := 0
	for _, s := range stars {
		if s == 3 {
			count++
		}
	}
	return count >= 10
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/game/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/game
git commit -m "feat(game): stars, combo, xp/levels, badges"
```

---

### Task 12: Progress persistence (v2 with v1 backup)

**Files:**
- Create: `internal/progress/progress.go`
- Test: `internal/progress/progress_test.go`

- [ ] **Step 1: Write the failing test**

```go
package progress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "progress.json")
	p := New()
	p.XP = 340
	p.Level = 2
	p.Stars["a1l1c1"] = 3
	p.MarkCompleted("a1l1c1")
	p.AddBadge("First Steps")
	p.LastLesson = "act1-01-the-two-stances"
	if err := Save(path, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got := Load(path)
	if got.XP != 340 || got.Level != 2 || got.Stars["a1l1c1"] != 3 ||
		!got.IsCompleted("a1l1c1") || !got.HasBadge("First Steps") ||
		got.LastLesson != "act1-01-the-two-stances" {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestLoadMissingFileStartsFresh(t *testing.T) {
	p := Load(filepath.Join(t.TempDir(), "nope.json"))
	if p.Version != 2 || p.XP != 0 || p.Level != 1 || p.Stars == nil {
		t.Errorf("fresh progress malformed: %+v", p)
	}
}

func TestLoadV1FileBacksUpAndStartsFresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	v1 := []byte(`{"xp":120,"streak":3,"completed_challenges":["c1"]}`)
	if err := os.WriteFile(path, v1, 0o644); err != nil {
		t.Fatal(err)
	}
	p := Load(path)
	if p.XP != 0 {
		t.Errorf("v1 file must not load as v2, got XP %d", p.XP)
	}
	if _, err := os.Stat(path + ".v1.bak"); err != nil {
		t.Errorf("expected backup file: %v", err)
	}
}

func TestLoadCorruptFileStartsFresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	os.WriteFile(path, []byte("{not json"), 0o644)
	p := Load(path)
	if p.Version != 2 {
		t.Errorf("corrupt file must yield fresh progress: %+v", p)
	}
}

func TestDoubleCompleteIsIdempotent(t *testing.T) {
	p := New()
	p.MarkCompleted("x")
	p.MarkCompleted("x")
	if len(p.Completed) != 1 {
		t.Errorf("Completed = %v, want single entry", p.Completed)
	}
	p.AddBadge("B")
	p.AddBadge("B")
	if len(p.Badges) != 1 {
		t.Errorf("Badges = %v, want single entry", p.Badges)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/progress/ -v`
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Implement — `internal/progress/progress.go`**

```go
// Package progress persists the player's journey to a versioned JSON file.
package progress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

const Version = 2

type Progress struct {
	Version    int            `json:"version"`
	XP         int            `json:"xp"`
	Level      int            `json:"level"`
	Stars      map[string]int `json:"stars"`
	Completed  []string       `json:"completed"`
	Badges     []string       `json:"badges"`
	LastLesson string         `json:"lastLesson"`
}

func New() *Progress {
	return &Progress{Version: Version, Level: 1, Stars: map[string]int{}}
}

func (p *Progress) IsCompleted(id string) bool { return slices.Contains(p.Completed, id) }

func (p *Progress) MarkCompleted(id string) {
	if !p.IsCompleted(id) {
		p.Completed = append(p.Completed, id)
	}
}

func (p *Progress) HasBadge(name string) bool { return slices.Contains(p.Badges, name) }

func (p *Progress) AddBadge(name string) {
	if !p.HasBadge(name) {
		p.Badges = append(p.Badges, name)
	}
}

// DefaultPath is ~/.nvim-quest/progress.json.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "progress.json"
	}
	return filepath.Join(home, ".nvim-quest", "progress.json")
}

// Load never fails: missing, corrupt, or old-version files yield a fresh
// Progress (old files are kept as <path>.v1.bak).
func Load(path string) *Progress {
	data, err := os.ReadFile(path)
	if err != nil {
		return New()
	}
	var p Progress
	if err := json.Unmarshal(data, &p); err != nil || p.Version != Version {
		_ = os.Rename(path, path+".v1.bak")
		return New()
	}
	if p.Stars == nil {
		p.Stars = map[string]int{}
	}
	if p.Level < 1 {
		p.Level = 1
	}
	return &p
}

func Save(path string, p *Progress) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/progress/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/progress
git commit -m "feat(progress): versioned persistence with v1 backup"
```

---

### Task 13: UI — the complete game (title, stats, map, room, boss, results)

The UI package is one coherent state machine; splitting it across tasks would force throwaway stubs. It's built here as one task with many small steps. Tests are model-level (Update/transition logic); views are exercised by compilation and manual play in Task 14.

**Files:**
- Create: `internal/ui/keys.go`, `internal/ui/styles.go`, `internal/ui/app.go`, `internal/ui/title.go`, `internal/ui/worldmap.go`, `internal/ui/room.go`, `internal/ui/results.go`
- Test: `internal/ui/app_test.go`

- [ ] **Step 1: Write the failing tests — `internal/ui/app_test.go`**

```go
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

func TestTitleNavigation(t *testing.T) {
	m := newTestModel(t)
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v`
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Create `internal/ui/keys.go`**

```go
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
```

- [ ] **Step 4: Create `internal/ui/styles.go`**

```go
package ui

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Primary lipgloss.Color
	Accent  lipgloss.Color
}

// Per-act palettes: dojo greens, crypt embers, neon grid.
var palettes = map[int]Palette{
	1: {Primary: lipgloss.Color("114"), Accent: lipgloss.Color("230")},
	2: {Primary: lipgloss.Color("214"), Accent: lipgloss.Color("203")},
	3: {Primary: lipgloss.Color("213"), Accent: lipgloss.Color("51")},
}

func paletteFor(act int) Palette {
	if p, ok := palettes[act]; ok {
		return p
	}
	return palettes[1]
}

func actName(act int) string {
	switch act {
	case 1:
		return "THE CURSOR DOJO"
	case 2:
		return "THE MOTION CRYPTS"
	case 3:
		return "THE NEON GRID"
	}
	return ""
}

var (
	cursorStyle  = lipgloss.NewStyle().Reverse(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	dangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)
```

- [ ] **Step 5: Create `internal/ui/app.go`**

```go
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
	return Model{
		width: 80, height: 24,
		lessons: lessons, prog: prog, savePath: savePath,
		combo: 1, actHeartsLost: map[int]bool{},
	}
}

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
```

- [ ] **Step 6: Create `internal/ui/title.go`**

```go
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
	case "j", "down":
		if m.menuIdx < len(menuItems)-1 {
			m.menuIdx++
		}
	case "k", "up":
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
```

- [ ] **Step 7: Create `internal/ui/worldmap.go`**

```go
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
	case "j", "down":
		if m.mapIdx < len(m.lessons)-1 {
			m.mapIdx++
		}
	case "k", "up":
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
```

- [ ] **Step 8: Create `internal/ui/room.go`**

```go
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nvim-quest/internal/content"
	"nvim-quest/internal/engine"
	"nvim-quest/internal/game"
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
	m.resFailed, m.resWasBoss = false, false
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
	b.WriteString(ch.Intro + "\n\n")
	b.WriteString(m.renderBuffer() + "\n\n")
	b.WriteString(m.renderHUD(ch) + "\n")
	if m.showHint {
		b.WriteString(successStyle.Render("hint: "+ch.Hint) + "\n")
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
```

- [ ] **Step 9: Create `internal/ui/results.go`**

```go
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
		b.WriteString(l.Boss.Name + " survives... this time.\n\n")
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
```

- [ ] **Step 10: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v`
Expected: PASS (all seven UI tests)

Then: `go test ./...`
Expected: PASS across all packages

- [ ] **Step 11: Commit**

```bash
git add internal/ui
git commit -m "feat(ui): five-screen game with real-time rooms, bosses, results"
```

---

### Task 14: CLI wiring, README, final verification

**Files:**
- Modify: `cmd/root.go` (replace placeholder RunE)
- Create: `cmd/stats.go`, `cmd/reset.go`
- Modify: `README.md`

- [ ] **Step 1: Wire the game into `cmd/root.go` (replace the whole file)**

```go
package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"nvim-quest/internal/content"
	"nvim-quest/internal/progress"
	"nvim-quest/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-quest",
	Short: "Learn Neovim through an epic three-act terminal quest",
	RunE: func(cmd *cobra.Command, args []string) error {
		lessons, err := content.All()
		if err != nil {
			return err
		}
		path := progress.DefaultPath()
		prog := progress.Load(path)
		p := tea.NewProgram(ui.New(lessons, prog, path), tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Create `cmd/stats.go`**

```go
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"nvim-quest/internal/progress"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your journey stats",
	Run: func(cmd *cobra.Command, args []string) {
		p := progress.Load(progress.DefaultPath())
		stars := 0
		for _, s := range p.Stars {
			stars += s
		}
		fmt.Printf("level %d · %d XP\n", p.Level, p.XP)
		fmt.Printf("rooms cleared: %d · stars: %d\n", len(p.Completed), stars)
		if len(p.Badges) > 0 {
			fmt.Println("badges: " + strings.Join(p.Badges, ", "))
		}
		if p.LastLesson != "" {
			fmt.Println("last lesson: " + p.LastLesson)
		}
	},
}

func init() { rootCmd.AddCommand(statsCmd) }
```

- [ ] **Step 3: Create `cmd/reset.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nvim-quest/internal/progress"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Erase all progress and start the journey anew",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("This erases ALL progress. Type 'yes' to confirm: ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "yes" {
			fmt.Println("aborted")
			return nil
		}
		if err := os.Remove(progress.DefaultPath()); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Println("progress erased — the journey begins again")
		return nil
	},
}

func init() { rootCmd.AddCommand(resetCmd) }
```

- [ ] **Step 4: Rewrite `README.md`**

```markdown
# nvim-quest

Learn Neovim through an epic three-act terminal quest — no Neovim required.

Real Vim keystrokes, instantly. Press `d` then `w` and the word vanishes,
exactly like the real editor. Earn stars by solving rooms under par (golf!),
build combo streaks, guard your hearts, and beat the clock in boss fights.

## The journey

- **Act I · The Cursor Dojo** — modes, `hjkl`, `w`/`b`, `0` `$` `gg` `G`
- **Act II · The Motion Crypts** — `i`/`a`, `x` `dw` `dd`, `yy` `p`, `cw` `cc`
- **Act III · The Neon Grid** — `/` search and `n`, count prefixes (`4w`, `3dd`)

Each act ends in a timed boss fight.

## Play

```sh
go run .            # start the game
go run . stats      # progress summary
go run . reset      # wipe progress
```

Or install it: `go install .` then run `nvim-quest` from anywhere
(lessons are embedded — no working-directory requirement).

Progress saves to `~/.nvim-quest/progress.json`.

## Develop

```sh
go test ./...
```

Lessons are JSON files in `assets/lessons/` — add one and the content
integrity test validates it automatically.

Curriculum inspired by
[Learn-Vim-and-NeoVim](https://github.com/rcallaby/Learn-Vim-and-NeoVim).

## Roadmap

- **Solid:** `e`, `o`/`O`, `P`, `ciw`, `f`/`t`, `:%s/old/new/g`, undo lesson
- **Everything:** visual mode, visual block, marks, registers, macros
```

- [ ] **Step 5: Full verification**

```bash
gofmt -l .            # expected: no output
go vet ./...          # expected: no findings
go test ./...         # expected: all packages PASS
go build ./...        # expected: clean build
```

- [ ] **Step 6: Manual smoke test**

Run `go run .` in a real terminal and verify, in order: title screen renders with menu →
World Map shows Act I unlocked / Acts II–III locked → enter "The Two Stances" →
press `z` (heart drops, no crash) → press `i`, type `I `, see instant buffer update and
the green clear flash → results show ⭐⭐⭐ and +50 XP → enter advances to room 2 →
esc returns to map → quit with `q` from title. Then `go run . stats` prints the XP.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: wire game into CLI, stats/reset commands, new README"
```

---

## Spec coverage checklist

| Spec section | Tasks |
|---|---|
| Real-time keystroke engine, `Press(key) → Event` | 2–7 |
| Operators dw/dd/cw/cc, counts, undo, yank/paste, search | 4–6 |
| Six goal validators | 8 |
| go:embed content, lesson/boss schema | 1, 9, 10 |
| 10 lessons + 3 bosses across three themed acts | 9, 10 |
| Stars/par, combo, hearts, XP/levels, badges | 11, 13 |
| Boss timer (tea.Tick), retry on failure | 13 |
| Progress v2 + v1 backup, save-failure warning | 12, 13 |
| Five screens, world map unlock flow, per-act palettes | 13 |
| Title/stats/reset CLI, 80×24 guard, README | 13, 14 |
```




