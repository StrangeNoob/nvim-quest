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

func (s *Simulator) deleteChar() {
	line := s.line()
	if len(line) == 0 || s.Cursor.Col >= len(line) {
		return
	}
	s.setLine(line[:s.Cursor.Col] + line[s.Cursor.Col+1:])
	s.clampCol()
}

func (s *Simulator) applyOperator(key string) Event {
	// Text object: d/c followed by "i" arms the inner-object prefix; the next key
	// (e.g. w) completes it as diw/ciw.
	if !s.pendingInner && (s.pendingOp == "d" || s.pendingOp == "c") && key == "i" {
		s.pendingInner = true
		return Event{EvPending}
	}
	if s.pendingInner {
		op := s.pendingOp
		s.clearPending()
		if key == "w" {
			s.snapshot()
			s.deleteInnerWord()
			if op == "c" {
				// ciw: stay at the word's start (may be end-of-line, a valid
				// insert position) and enter insert; do NOT normal-mode clamp.
				s.Mode = ModeInsert
				return Event{EvModeChanged}
			}
			s.clampCol() // diw: settle the cursor on a valid normal-mode column
			return Event{EvEdited}
		}
		return Event{EvInvalid}
	}
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

func (s *Simulator) currentSnapshot() snapshot {
	buf := make([]string, len(s.Buffer))
	copy(buf, s.Buffer)
	return snapshot{buffer: buf, cursor: s.Cursor}
}

func (s *Simulator) snapshot() {
	s.undo = append(s.undo, s.currentSnapshot())
	s.redo = nil // a new change invalidates the redo history
}

func (s *Simulator) applyUndo() Event {
	if len(s.undo) == 0 {
		return Event{EvNone}
	}
	s.redo = append(s.redo, s.currentSnapshot())
	last := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	s.Buffer = last.buffer
	s.Cursor = last.cursor
	return Event{EvEdited}
}

func (s *Simulator) applyRedo() Event {
	if len(s.redo) == 0 {
		return Event{EvNone}
	}
	s.undo = append(s.undo, s.currentSnapshot())
	last := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]
	s.Buffer = last.buffer
	s.Cursor = last.cursor
	return Event{EvEdited}
}

func (s *Simulator) paste() Event { return s.pasteAt(s.Cursor.Row + 1) }

func (s *Simulator) pasteBefore() Event { return s.pasteAt(s.Cursor.Row) }

// pasteAt inserts the yank register's lines starting at row `at`, moving the
// cursor to the first pasted line. No-op on an empty register.
func (s *Simulator) pasteAt(at int) Event {
	if len(s.yank) == 0 {
		return Event{EvNone}
	}
	s.snapshot()
	out := make([]string, 0, len(s.Buffer)+len(s.yank))
	out = append(out, s.Buffer[:at]...)
	out = append(out, s.yank...)
	out = append(out, s.Buffer[at:]...)
	s.Buffer = out
	s.Cursor = Pos{at, 0}
	return Event{EvEdited}
}

// openLine inserts a new empty line below (or above) the cursor's line and
// enters Insert mode, like Vim's o / O.
func (s *Simulator) openLine(below bool) {
	s.snapshot()
	at := s.Cursor.Row
	if below {
		at++
	}
	out := make([]string, 0, len(s.Buffer)+1)
	out = append(out, s.Buffer[:at]...)
	out = append(out, "")
	out = append(out, s.Buffer[at:]...)
	s.Buffer = out
	s.Cursor = Pos{at, 0}
	s.Mode = ModeInsert
}

// deleteInnerWord removes the whole space-delimited word the cursor sits on
// (no trailing space), leaving the cursor at the word's start (diw / ciw).
func (s *Simulator) deleteInnerWord() {
	line := s.line()
	col := s.Cursor.Col
	if col >= len(line) || line[col] == ' ' {
		return
	}
	start := col
	for start > 0 && line[start-1] != ' ' {
		start--
	}
	end := col
	for end < len(line) && line[end] != ' ' {
		end++
	}
	s.setLine(line[:start] + line[end:])
	s.Cursor.Col = start // caller clamps for diw; ciw keeps this as the insert point
}
