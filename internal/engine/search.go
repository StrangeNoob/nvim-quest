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
