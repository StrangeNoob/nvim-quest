package engine

import "testing"

func TestMotions(t *testing.T) {
	tests := []struct {
		name   string
		buffer []string
		cursor Pos
		keys   []string
		want   Pos
	}{
		{"h moves left", []string{"abc"}, Pos{0, 1}, []string{"h"}, Pos{0, 0}},
		{"h stops at col 0", []string{"abc"}, Pos{0, 0}, []string{"h"}, Pos{0, 0}},
		{"l moves right", []string{"abc"}, Pos{0, 0}, []string{"l"}, Pos{0, 1}},
		{"l stops at line end", []string{"ab"}, Pos{0, 1}, []string{"l"}, Pos{0, 1}},
		{"j moves down and clamps col", []string{"abcdef", "ab"}, Pos{0, 5}, []string{"j"}, Pos{1, 1}},
		{"j stops at last row", []string{"a", "b"}, Pos{1, 0}, []string{"j"}, Pos{1, 0}},
		{"k moves up", []string{"a", "b"}, Pos{1, 0}, []string{"k"}, Pos{0, 0}},
		{"0 jumps to line start", []string{"hello"}, Pos{0, 4}, []string{"0"}, Pos{0, 0}},
		{"$ jumps to line end", []string{"hello"}, Pos{0, 0}, []string{"$"}, Pos{0, 4}},
		{"gg jumps to top", []string{"a", "b", "c"}, Pos{2, 0}, []string{"g", "g"}, Pos{0, 0}},
		{"G jumps to bottom", []string{"a", "b", "c"}, Pos{0, 0}, []string{"G"}, Pos{2, 0}},
		{"w jumps to next word", []string{"foo bar baz"}, Pos{0, 0}, []string{"w"}, Pos{0, 4}},
		{"w crosses to next line", []string{"foo", "bar"}, Pos{0, 0}, []string{"w"}, Pos{1, 0}},
		{"w at last word stays on line end", []string{"foo bar"}, Pos{0, 4}, []string{"w"}, Pos{0, 6}},
		{"b jumps to prev word start", []string{"foo bar"}, Pos{0, 4}, []string{"b"}, Pos{0, 0}},
		{"b inside word jumps to its start", []string{"foo bar"}, Pos{0, 6}, []string{"b"}, Pos{0, 4}},
		{"b crosses to prev line", []string{"foo", "bar"}, Pos{1, 0}, []string{"b"}, Pos{0, 0}},
		{"count 3w", []string{"a b c d"}, Pos{0, 0}, []string{"3", "w"}, Pos{0, 6}},
		{"count 2j", []string{"a", "b", "c"}, Pos{0, 0}, []string{"2", "j"}, Pos{2, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
			}
			if s.Cursor != tt.want {
				t.Errorf("cursor = %v, want %v", s.Cursor, tt.want)
			}
		})
	}
}

func TestNewCopiesBuffer(t *testing.T) {
	src := []string{"abc"}
	s := New(src, Pos{0, 0})
	src[0] = "mutated"
	if s.Buffer[0] != "abc" {
		t.Errorf("New must copy the buffer, got %q", s.Buffer[0])
	}
}
