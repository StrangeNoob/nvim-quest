package engine

import "testing"

func TestGoals(t *testing.T) {
	tests := []struct {
		name  string
		goal  Goal
		setup func() *Simulator
		want  bool
	}{
		{"cursorOnWord hit", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 6}) }, true},
		{"cursorOnWord middle of word counts", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 8}) }, true},
		{"cursorOnWord miss", Goal{Type: "cursorOnWord", Word: "temple"},
			func() *Simulator { return New([]string{"the temple gate"}, Pos{0, 0}) }, false},
		{"bufferEquals hit", Goal{Type: "bufferEquals", Lines: []string{"a", "b"}},
			func() *Simulator { return New([]string{"a", "b"}, Pos{0, 0}) }, true},
		{"bufferEquals miss", Goal{Type: "bufferEquals", Lines: []string{"a"}},
			func() *Simulator { return New([]string{"a", "b"}, Pos{0, 0}) }, false},
		{"lineDeleted hit", Goal{Type: "lineDeleted", Line: "gone"},
			func() *Simulator { return New([]string{"keep"}, Pos{0, 0}) }, true},
		{"lineDeleted miss", Goal{Type: "lineDeleted", Line: "gone"},
			func() *Simulator { return New([]string{"keep", "gone"}, Pos{0, 0}) }, false},
		{"wordDeleted hit", Goal{Type: "wordDeleted", Word: "cursed"},
			func() *Simulator { return New([]string{"the goblin"}, Pos{0, 0}) }, true},
		{"wordDeleted miss", Goal{Type: "wordDeleted", Word: "cursed"},
			func() *Simulator { return New([]string{"the cursed goblin"}, Pos{0, 0}) }, false},
		{"containsText hit", Goal{Type: "containsText", Text: "I am ready"},
			func() *Simulator { return New([]string{"I am ready now"}, Pos{0, 0}) }, true},
		{"containsText miss", Goal{Type: "containsText", Text: "I am ready"},
			func() *Simulator { return New([]string{"am ready"}, Pos{0, 0}) }, false},
		{"searchMatchActive hit", Goal{Type: "searchMatchActive", Term: "backdoor"},
			func() *Simulator {
				s := New([]string{"a backdoor hides"}, Pos{0, 0})
				for _, k := range []string{"/", "b", "a", "c", "k", "d", "o", "o", "r", "enter"} {
					s.Press(k)
				}
				return s
			}, true},
		{"searchMatchActive miss without search", Goal{Type: "searchMatchActive", Term: "backdoor"},
			func() *Simulator { return New([]string{"a backdoor hides"}, Pos{0, 2}) }, false},
		{"unknown goal type is never met", Goal{Type: "bogus"},
			func() *Simulator { return New([]string{"a"}, Pos{0, 0}) }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.goal.Met(tt.setup()); got != tt.want {
				t.Errorf("Met() = %v, want %v", got, tt.want)
			}
		})
	}
}
