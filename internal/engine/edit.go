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
