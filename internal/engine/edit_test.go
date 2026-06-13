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
		{"dw at end of line does not join next line", []string{"foo", "bar"}, Pos{0, 0},
			[]string{"d", "w"}, []string{"", "bar"}, Pos{0, 0}, ModeNormal},
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
