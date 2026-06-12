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
