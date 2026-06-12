package engine

import (
	"testing"

	"nvim-quest/internal/lessons"
)

func TestValidators(t *testing.T) {
	tests := []struct {
		name      string
		simulator *Simulator
		success   lessons.Success
	}{
		{"cursor on word", NewSimulator([]string{"go TARGET"}, Cursor{Col: 3}), lessons.Success{CursorOnWord: "TARGET"}},
		{"buffer equals", NewSimulator([]string{"done"}, Cursor{}), lessons.Success{BufferEquals: []string{"done"}}},
		{"line deleted", NewSimulator([]string{"keep"}, Cursor{}), lessons.Success{LineDeleted: "remove"}},
		{"word deleted", NewSimulator([]string{"keep text"}, Cursor{}), lessons.Success{WordDeleted: "extra"}},
		{"contains text", NewSimulator([]string{"hello Vim"}, Cursor{}), lessons.Success{ContainsText: "Vim"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !Validate(test.simulator, lessons.Challenge{Success: test.success}) {
				t.Fatal("expected validation success")
			}
		})
	}
}

func TestSearchMatchValidator(t *testing.T) {
	simulator := NewSimulator([]string{"find needle"}, Cursor{})
	command, _ := ParseCommand("/needle")
	if err := simulator.Apply(command); err != nil {
		t.Fatal(err)
	}
	challenge := lessons.Challenge{Success: lessons.Success{SearchMatchActive: "needle"}}
	if !Validate(simulator, challenge) {
		t.Fatal("expected active search to validate")
	}
}
