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
		return g.Word != "" && s.WordAtCursor() == g.Word
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
		if g.Term == "" {
			return false
		}
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
