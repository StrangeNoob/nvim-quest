// Package engine is a small real-time Vim emulator: feed it keystrokes,
// it mutates buffer/cursor/mode state and reports what happened.
package engine

import "strconv"

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

// Pending returns the visible pending state ("d", "3", "3d", "g") for the HUD.
func (s *Simulator) Pending() string {
	out := ""
	if s.pendingOp != "" {
		if s.opCount > 1 {
			out += strconv.Itoa(s.opCount)
		}
		out += s.pendingOp
	} else if s.pendingCount > 0 {
		out += strconv.Itoa(s.pendingCount)
	}
	if s.pendingG {
		out += "g"
	}
	return out
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
