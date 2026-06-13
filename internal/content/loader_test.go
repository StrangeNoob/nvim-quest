package content

import "testing"

func TestAllLoadsAndValidates(t *testing.T) {
	lessons, err := All()
	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	if len(lessons) == 0 {
		t.Fatal("no lessons loaded")
	}
	seen := map[string]bool{}
	for i, l := range lessons {
		if l.ID == "" || l.Title == "" {
			t.Errorf("lesson %d missing id/title", i)
		}
		if l.Act < 1 || l.Act > 3 {
			t.Errorf("%s: act %d out of range", l.ID, l.Act)
		}
		if i > 0 {
			prev := lessons[i-1]
			if l.Act < prev.Act || (l.Act == prev.Act && l.Order <= prev.Order) {
				t.Errorf("%s not sorted after %s", l.ID, prev.ID)
			}
		}
		if len(l.Challenges) == 0 {
			t.Errorf("%s has no challenges", l.ID)
		}
		for _, ch := range append([]Challenge{}, l.Challenges...) {
			if seen[ch.ID] {
				t.Errorf("duplicate challenge id %s", ch.ID)
			}
			seen[ch.ID] = true
			validateChallenge(t, l.ID, ch, true)
		}
		if l.Boss != nil {
			if l.Boss.TimeLimitSec < 30 {
				t.Errorf("%s boss time limit too low", l.ID)
			}
			if l.Boss.XP <= 0 {
				t.Errorf("%s boss has no xp", l.ID)
			}
			if len(l.Boss.Steps) == 0 {
				t.Errorf("%s boss has no steps", l.ID)
			}
			for _, st := range l.Boss.Steps {
				validateChallenge(t, l.ID+":boss", st, false)
			}
		}
	}
	if len(lessons) != 10 {
		t.Errorf("expected 10 lessons, got %d", len(lessons))
	}
	bosses := 0
	lastInAct := map[int]string{}
	for _, l := range lessons {
		lastInAct[l.Act] = l.ID
		if l.Boss != nil {
			bosses++
		}
	}
	if bosses != 3 {
		t.Errorf("expected 3 bosses, got %d", bosses)
	}
	for act, id := range lastInAct {
		for _, l := range lessons {
			if l.ID == id && l.Boss == nil {
				t.Errorf("act %d final lesson %s must carry the boss", act, id)
			}
		}
	}
}

func validateChallenge(t *testing.T, owner string, ch Challenge, needsParXP bool) {
	t.Helper()
	if len(ch.Buffer) == 0 {
		t.Errorf("%s/%s: empty buffer", owner, ch.ID)
	}
	row, col := ch.Cursor[0], ch.Cursor[1]
	if row < 0 || row >= len(ch.Buffer) || col < 0 || (len(ch.Buffer) > row && col > len(ch.Buffer[row])) {
		t.Errorf("%s/%s: cursor %v out of bounds", owner, ch.ID, ch.Cursor)
	}
	valid := false
	for _, gt := range goalTypes() {
		if ch.Goal.Type == gt {
			valid = true
		}
	}
	if !valid {
		t.Errorf("%s/%s: unknown goal type %q", owner, ch.ID, ch.Goal.Type)
	}
	if needsParXP {
		if ch.Par < 1 {
			t.Errorf("%s/%s: par must be >= 1", owner, ch.ID)
		}
		if ch.XP <= 0 {
			t.Errorf("%s/%s: xp must be > 0", owner, ch.ID)
		}
	}
}
