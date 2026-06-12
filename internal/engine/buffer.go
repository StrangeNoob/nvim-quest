package engine

import "strings"

type Buffer struct {
	Lines  []string
	Cursor Cursor
}

func NewBuffer(lines []string, cursor Cursor) Buffer {
	copied := append([]string(nil), lines...)
	if len(copied) == 0 {
		copied = []string{""}
	}
	buffer := Buffer{Lines: copied, Cursor: cursor}
	buffer.Clamp()
	return buffer
}

func (b *Buffer) Clamp() {
	if len(b.Lines) == 0 {
		b.Lines = []string{""}
	}
	b.Cursor.Row = min(max(b.Cursor.Row, 0), len(b.Lines)-1)
	line := b.Lines[b.Cursor.Row]
	maxCol := max(len(line)-1, 0)
	b.Cursor.Col = min(max(b.Cursor.Col, 0), maxCol)
}

func (b Buffer) Clone() Buffer {
	return NewBuffer(b.Lines, b.Cursor)
}

func (b Buffer) CurrentLine() string {
	return b.Lines[b.Cursor.Row]
}

func (b Buffer) Text() string {
	return strings.Join(b.Lines, "\n")
}
