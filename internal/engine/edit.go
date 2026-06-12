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
