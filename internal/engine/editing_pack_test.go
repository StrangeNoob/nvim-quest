package engine

import (
	"slices"
	"testing"
)

func TestOpenLineAndPasteBefore(t *testing.T) {
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantBuffer []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"o opens below and inserts", []string{"light"}, Pos{0, 0},
			[]string{"o", "d", "a", "r", "k"}, []string{"light", "dark"}, Pos{1, 4}, ModeInsert},
		{"O opens above and inserts", []string{"second"}, Pos{0, 0},
			[]string{"O", "f", "i", "r", "s", "t"}, []string{"first", "second"}, Pos{0, 5}, ModeInsert},
		{"o below a middle line", []string{"a", "b", "c"}, Pos{1, 0},
			[]string{"o", "x"}, []string{"a", "b", "x", "c"}, Pos{2, 1}, ModeInsert},
		{"P pastes before the current line", []string{"top", "bottom"}, Pos{1, 0},
			[]string{"k", "y", "y", "j", "P"}, []string{"top", "top", "bottom"}, Pos{1, 0}, ModeNormal},
		{"P with empty register is a no-op", []string{"only"}, Pos{0, 0},
			[]string{"P"}, []string{"only"}, Pos{0, 0}, ModeNormal},
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

func TestInnerWord(t *testing.T) {
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantBuffer []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"diw deletes the word under the cursor", []string{"delete the cursed rune"}, Pos{0, 11},
			[]string{"d", "i", "w"}, []string{"delete the  rune"}, Pos{0, 11}, ModeNormal},
		{"diw from mid-word", []string{"foo bar baz"}, Pos{0, 5},
			[]string{"d", "i", "w"}, []string{"foo  baz"}, Pos{0, 4}, ModeNormal},
		{"ciw changes the inner word", []string{"the wrong word"}, Pos{0, 4},
			[]string{"c", "i", "w", "r", "i", "g", "h", "t"}, []string{"the right word"}, Pos{0, 9}, ModeInsert},
		{"ciw on the last word of a line inserts at the right offset", []string{"foo bar"}, Pos{0, 5},
			[]string{"c", "i", "w", "b", "a", "z"}, []string{"foo baz"}, Pos{0, 7}, ModeInsert},
		{"diw on a space is a no-op", []string{"a b"}, Pos{0, 1},
			[]string{"d", "i", "w"}, []string{"a b"}, Pos{0, 1}, ModeNormal},
		{"d then invalid object cancels", []string{"abc def"}, Pos{0, 0},
			[]string{"d", "i", "x"}, []string{"abc def"}, Pos{0, 0}, ModeNormal},
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

func TestRedo(t *testing.T) {
	t.Run("ctrl+r redoes an undone edit", func(t *testing.T) {
		s := New([]string{"echo"}, Pos{0, 0})
		s.Press("y")
		s.Press("y")
		s.Press("p") // ["echo","echo"]
		s.Press("u") // back to ["echo"]
		if !slices.Equal(s.Buffer, []string{"echo"}) {
			t.Fatalf("after undo buffer = %q, want [echo]", s.Buffer)
		}
		s.Press("ctrl+r") // redo → ["echo","echo"]
		if !slices.Equal(s.Buffer, []string{"echo", "echo"}) {
			t.Errorf("after redo buffer = %q, want [echo echo]", s.Buffer)
		}
	})

	t.Run("a new edit clears the redo history", func(t *testing.T) {
		s := New([]string{"a"}, Pos{0, 0})
		s.Press("x") // delete 'a' → [""]
		s.Press("u") // undo → ["a"], redo has [""]
		s.Press("o") // new edit (open line) clears redo
		s.Press("esc")
		s.Press("ctrl+r") // nothing to redo
		if !slices.Equal(s.Buffer, []string{"a", ""}) {
			t.Errorf("redo after a new edit must be a no-op; buffer = %q", s.Buffer)
		}
	})

	t.Run("ctrl+r with empty redo stack is a no-op", func(t *testing.T) {
		s := New([]string{"ab"}, Pos{0, 0})
		if ev := s.Press("ctrl+r"); ev.Kind != EvNone {
			t.Errorf("redo on empty stack = %v, want EvNone", ev.Kind)
		}
	})
}
