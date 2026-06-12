package engine

import (
	"slices"
	"testing"
)

func TestCursorMovement(t *testing.T) {
	simulator := NewSimulator([]string{"abc", "de"}, Cursor{})
	applyAll(t, simulator, "l", "l", "j")
	if simulator.Buffer.Cursor != (Cursor{Row: 1, Col: 1}) {
		t.Fatalf("cursor = %+v", simulator.Buffer.Cursor)
	}
	applyAll(t, simulator, "0", "k", "$")
	if simulator.Buffer.Cursor != (Cursor{Row: 0, Col: 2}) {
		t.Fatalf("cursor = %+v", simulator.Buffer.Cursor)
	}
	applyAll(t, simulator, "G", "gg")
	if simulator.Buffer.Cursor != (Cursor{}) {
		t.Fatalf("cursor = %+v", simulator.Buffer.Cursor)
	}
}

func TestWordMovement(t *testing.T) {
	simulator := NewSimulator([]string{"one two three"}, Cursor{})
	applyAll(t, simulator, "w", "w")
	if simulator.Buffer.Cursor.Col != 8 {
		t.Fatalf("col = %d", simulator.Buffer.Cursor.Col)
	}
	applyAll(t, simulator, "b")
	if simulator.Buffer.Cursor.Col != 4 {
		t.Fatalf("col = %d", simulator.Buffer.Cursor.Col)
	}
}

func TestDeleteWordAndUndo(t *testing.T) {
	simulator := NewSimulator([]string{"keep extra text"}, Cursor{Col: 5})
	applyAll(t, simulator, "dw")
	if got := simulator.Buffer.CurrentLine(); got != "keep text" {
		t.Fatalf("line = %q", got)
	}
	applyAll(t, simulator, "u")
	if got := simulator.Buffer.CurrentLine(); got != "keep extra text" {
		t.Fatalf("undo line = %q", got)
	}
}

func TestDeleteLine(t *testing.T) {
	simulator := NewSimulator([]string{"one", "remove", "three"}, Cursor{Row: 1})
	applyAll(t, simulator, "dd")
	if !slices.Equal(simulator.Buffer.Lines, []string{"one", "three"}) {
		t.Fatalf("lines = %#v", simulator.Buffer.Lines)
	}
}

func TestYankAndPaste(t *testing.T) {
	simulator := NewSimulator([]string{"copy", "end"}, Cursor{})
	applyAll(t, simulator, "yy", "p")
	if !slices.Equal(simulator.Buffer.Lines, []string{"copy", "copy", "end"}) {
		t.Fatalf("lines = %#v", simulator.Buffer.Lines)
	}
}

func TestInsertCommands(t *testing.T) {
	insert, err := ParseCommand("ihello <Esc>")
	if err != nil || insert.Name != "i" || insert.Text != "hello " {
		t.Fatalf("insert = %+v, err = %v", insert, err)
	}
	appendCommand, err := ParseCommand("a!<Esc>")
	if err != nil || appendCommand.Name != "a" || appendCommand.Text != "!" {
		t.Fatalf("append = %+v, err = %v", appendCommand, err)
	}

	simulator := NewSimulator([]string{"Vim"}, Cursor{Col: 2})
	if err := simulator.Apply(appendCommand); err != nil {
		t.Fatal(err)
	}
	if got := simulator.Buffer.CurrentLine(); got != "Vim!" {
		t.Fatalf("line = %q", got)
	}
}

func TestSearch(t *testing.T) {
	simulator := NewSimulator([]string{"start", "find needle", "needle again"}, Cursor{})
	applyAll(t, simulator, "/needle")
	if !simulator.SearchActive || simulator.Buffer.Cursor != (Cursor{Row: 1, Col: 5}) {
		t.Fatalf("search state = %v, cursor = %+v", simulator.SearchActive, simulator.Buffer.Cursor)
	}
	applyAll(t, simulator, "n")
	if simulator.Buffer.Cursor != (Cursor{Row: 2, Col: 0}) {
		t.Fatalf("cursor = %+v", simulator.Buffer.Cursor)
	}
}

func applyAll(t *testing.T, simulator *Simulator, inputs ...string) {
	t.Helper()
	for _, input := range inputs {
		command, err := ParseCommand(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := simulator.Apply(command); err != nil {
			t.Fatal(err)
		}
	}
}
