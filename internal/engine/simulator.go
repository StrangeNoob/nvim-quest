package engine

import (
	"fmt"
	"strings"
	"unicode"
)

type snapshot struct {
	buffer Buffer
}

type Simulator struct {
	Buffer       Buffer
	YankRegister string
	SearchTerm   string
	SearchActive bool
	history      []snapshot
}

func NewSimulator(lines []string, cursor Cursor) *Simulator {
	return &Simulator{Buffer: NewBuffer(lines, cursor)}
}

func (s *Simulator) Apply(command Command) error {
	switch command.Name {
	case "h":
		s.Buffer.Cursor.Col--
	case "l":
		s.Buffer.Cursor.Col++
	case "j":
		s.Buffer.Cursor.Row++
	case "k":
		s.Buffer.Cursor.Row--
	case "0":
		s.Buffer.Cursor.Col = 0
	case "$":
		s.Buffer.Cursor.Col = len(s.Buffer.CurrentLine()) - 1
	case "gg":
		s.Buffer.Cursor.Row, s.Buffer.Cursor.Col = 0, 0
	case "G":
		s.Buffer.Cursor.Row, s.Buffer.Cursor.Col = len(s.Buffer.Lines)-1, 0
	case "w":
		s.moveWordForward()
	case "b":
		s.moveWordBackward()
	case "x":
		s.save()
		s.deleteCharacter()
	case "dw":
		s.save()
		s.deleteWord()
	case "dd":
		s.save()
		s.deleteLine()
	case "yy":
		s.YankRegister = s.Buffer.CurrentLine()
	case "p":
		if s.YankRegister == "" {
			return fmt.Errorf("nothing has been yanked yet")
		}
		s.save()
		s.pasteLine()
	case "i":
		s.save()
		s.insert(command.Text, false)
	case "a":
		s.save()
		s.insert(command.Text, true)
	case "u":
		if len(s.history) == 0 {
			return fmt.Errorf("nothing to undo")
		}
		last := s.history[len(s.history)-1]
		s.history = s.history[:len(s.history)-1]
		s.Buffer = last.buffer
	case "/":
		s.SearchTerm = command.Text
		if !s.findNext(true) {
			s.SearchActive = false
			return fmt.Errorf("search term %q not found", command.Text)
		}
	case "n":
		if s.SearchTerm == "" {
			return fmt.Errorf("start a search with /word first")
		}
		if !s.findNext(false) {
			return fmt.Errorf("no next match for %q", s.SearchTerm)
		}
	default:
		return fmt.Errorf("unsupported command %q", command.Name)
	}
	s.Buffer.Clamp()
	return nil
}

func (s *Simulator) save() {
	s.history = append(s.history, snapshot{buffer: s.Buffer.Clone()})
}

func (s *Simulator) deleteCharacter() {
	line := s.Buffer.CurrentLine()
	if len(line) == 0 {
		return
	}
	col := s.Buffer.Cursor.Col
	s.Buffer.Lines[s.Buffer.Cursor.Row] = line[:col] + line[col+1:]
}

func (s *Simulator) deleteWord() {
	line := s.Buffer.CurrentLine()
	start := min(s.Buffer.Cursor.Col, len(line))
	end := start
	for end < len(line) && unicode.IsSpace(rune(line[end])) {
		end++
	}
	for end < len(line) && !unicode.IsSpace(rune(line[end])) {
		end++
	}
	for end < len(line) && unicode.IsSpace(rune(line[end])) {
		end++
	}
	s.Buffer.Lines[s.Buffer.Cursor.Row] = line[:start] + line[end:]
}

func (s *Simulator) deleteLine() {
	row := s.Buffer.Cursor.Row
	s.Buffer.Lines = append(s.Buffer.Lines[:row], s.Buffer.Lines[row+1:]...)
	if len(s.Buffer.Lines) == 0 {
		s.Buffer.Lines = []string{""}
	}
}

func (s *Simulator) pasteLine() {
	row := s.Buffer.Cursor.Row + 1
	s.Buffer.Lines = append(s.Buffer.Lines, "")
	copy(s.Buffer.Lines[row+1:], s.Buffer.Lines[row:])
	s.Buffer.Lines[row] = s.YankRegister
	s.Buffer.Cursor.Row, s.Buffer.Cursor.Col = row, 0
}

func (s *Simulator) insert(text string, after bool) {
	line := s.Buffer.CurrentLine()
	index := min(s.Buffer.Cursor.Col, len(line))
	if after && len(line) > 0 {
		index++
	}
	s.Buffer.Lines[s.Buffer.Cursor.Row] = line[:index] + text + line[index:]
	if len(text) > 0 {
		s.Buffer.Cursor.Col = index + len(text) - 1
	}
}

func (s *Simulator) moveWordForward() {
	text := s.Buffer.Text()
	offset := s.offset()
	for offset < len(text) && !unicode.IsSpace(rune(text[offset])) {
		offset++
	}
	for offset < len(text) && unicode.IsSpace(rune(text[offset])) {
		offset++
	}
	s.setOffset(min(offset, len(text)-1))
}

func (s *Simulator) moveWordBackward() {
	text := s.Buffer.Text()
	offset := max(s.offset()-1, 0)
	for offset > 0 && unicode.IsSpace(rune(text[offset])) {
		offset--
	}
	for offset > 0 && !unicode.IsSpace(rune(text[offset-1])) {
		offset--
	}
	s.setOffset(offset)
}

func (s *Simulator) findNext(fromCursor bool) bool {
	text := s.Buffer.Text()
	start := s.offset()
	if !fromCursor {
		start++
	}
	if start > len(text) {
		return false
	}
	index := strings.Index(text[start:], s.SearchTerm)
	if index < 0 {
		return false
	}
	s.setOffset(start + index)
	s.SearchActive = true
	return true
}

func (s *Simulator) offset() int {
	offset := 0
	for row := 0; row < s.Buffer.Cursor.Row; row++ {
		offset += len(s.Buffer.Lines[row]) + 1
	}
	return offset + s.Buffer.Cursor.Col
}

func (s *Simulator) setOffset(offset int) {
	for row, line := range s.Buffer.Lines {
		if offset <= len(line) {
			s.Buffer.Cursor = Cursor{Row: row, Col: min(offset, max(len(line)-1, 0))}
			return
		}
		offset -= len(line) + 1
	}
	s.Buffer.Cursor = Cursor{Row: len(s.Buffer.Lines) - 1, Col: 0}
}
