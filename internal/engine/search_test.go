package engine

import "testing"

func TestSearch(t *testing.T) {
	buf := []string{"trace node1 active", "trace node2 active", "trace node3 active"}
	tests := []struct {
		name       string
		buffer     []string
		cursor     Pos
		keys       []string
		wantCursor Pos
		wantMode   Mode
	}{
		{"/term enter jumps to first match", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter"}, Pos{0, 6}, ModeNormal},
		{"n repeats forward", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter", "n"}, Pos{1, 6}, ModeNormal},
		{"n wraps around", buf, Pos{0, 0},
			[]string{"/", "n", "o", "d", "e", "enter", "n", "n", "n"}, Pos{0, 6}, ModeNormal},
		{"esc cancels search", buf, Pos{0, 2},
			[]string{"/", "z", "esc"}, Pos{0, 2}, ModeNormal},
		{"backspace edits query", buf, Pos{0, 0},
			[]string{"/", "n", "z", "backspace", "o", "d", "e", "enter"}, Pos{0, 6}, ModeNormal},
		{"no match stays put", buf, Pos{0, 2},
			[]string{"/", "q", "q", "enter"}, Pos{0, 2}, ModeNormal},
		{"n without prior search is no-op", buf, Pos{0, 2},
			[]string{"n"}, Pos{0, 2}, ModeNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.buffer, tt.cursor)
			for _, k := range tt.keys {
				s.Press(k)
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

// TestSearchWrapsToEarlierMatchOnStartRow locks in the intentional rows+1
// iteration in jumpToMatch: a forward search must wrap around and find a match
// that lies BEFORE the cursor on the cursor's own line. A naive `i < rows` loop
// would miss it, so this test guards against that regression.
func TestSearchWrapsToEarlierMatchOnStartRow(t *testing.T) {
	// Single line, two matches; cursor sits after the only-line's matches.
	s := New([]string{"foo bar foo"}, Pos{0, 8}) // on the second "foo"
	for _, k := range []string{"/", "f", "o", "o", "enter"} {
		s.Press(k)
	}
	// enter searches forward from col 9; the only remaining match is at col 0
	// after wrapping around the single line.
	if s.Cursor != (Pos{0, 0}) {
		t.Errorf("cursor = %v, want {0 0} (search must wrap to the earlier match)", s.Cursor)
	}
}

func TestSearchQueryVisible(t *testing.T) {
	s := New([]string{"abc"}, Pos{0, 0})
	s.Press("/")
	s.Press("a")
	s.Press("b")
	if s.SearchQuery != "ab" {
		t.Errorf("SearchQuery = %q, want %q", s.SearchQuery, "ab")
	}
	if s.Mode != ModeSearch {
		t.Errorf("mode = %v, want ModeSearch", s.Mode)
	}
}
