package content

import (
	"testing"

	"github.com/StrangeNoob/nvim-quest/internal/engine"
)

// TestEveryChallengeIsSolvable plays an authored optimal solution through the
// real engine for every regular challenge and asserts the goal is met within
// par, using the challenge's own AllowedKeys. This is the permanent guard that
// every lesson is completable — and it refutes static false positives about
// "unsolvable" buffers (e.g. the deliberate double-space in act2-05).
func TestEveryChallengeIsSolvable(t *testing.T) {
	// challenge id -> optimal key sequence (esc is free and not counted by the UI)
	solutions := map[string][]string{
		"a1l1c1":  {"i", "I", " "},
		"a1l1c2":  {"i", "i", "n"},
		"a1l2c1":  {"j", "l", "l", "l", "l", "l", "l", "l"},
		"a1l2c2":  {"k", "k", "l", "l", "l", "l"},
		"a1l3c1":  {"w", "w", "w", "w"},
		"a1l3c2":  {"b", "b", "b"},
		"a1l4c1":  {"G", "w"},
		"a1l4c2":  {"g", "g", "$"},
		"a2l5c1":  {"a", "o", "p", "e", "n"},
		"a2l5c2":  {"i", "o", "l", "d"},
		"a2l6c1":  {"x"},
		"a2l6c2":  {"d", "w"},
		"a2l6c3":  {"d", "d"},
		"a2l7c1":  {"y", "y", "p"},
		"a2l7c2":  {"y", "y", "j", "p"},
		"a2l7c3":  {"k", "y", "y", "j", "P"},
		"a2l8c1":  {"o", "d", "a", "r", "k"},
		"a2l8c2":  {"O", "f", "i", "r", "s", "t"},
		"a2l9c1":  {"c", "i", "w", "r", "i", "g", "h", "t"},
		"a2l9c2":  {"d", "i", "w"},
		"a2l10c1": {"d", "d"},
		"a2l10c2": {"y", "y", "p"},
		"a2l11c1": {"c", "w", "m", "o", "o", "n"},
		"a2l11c2": {"c", "c", "r", "e", "b", "o", "r", "n"},
		"a3l9c1":  {"/", "b", "a", "c", "k", "d", "o", "o", "r", "enter"},
		"a3l9c2":  {"/", "n", "o", "d", "e", "enter", "n", "n"},
		"a3l10c1": {"4", "w"},
		"a3l10c2": {"3", "d", "d"},
		"a3l10c3": {"4", "x"},
	}

	lessons, err := All()
	if err != nil {
		t.Fatalf("All() error: %v", err)
	}

	seen := map[string]bool{}
	for _, l := range lessons {
		for _, ch := range l.Challenges {
			keys, ok := solutions[ch.ID]
			if !ok {
				t.Errorf("%s: no solution authored for this challenge", ch.ID)
				continue
			}
			seen[ch.ID] = true

			s := engine.New(ch.Buffer, engine.Pos{Row: ch.Cursor[0], Col: ch.Cursor[1]})
			if len(ch.AllowedKeys) > 0 {
				allow := map[string]bool{}
				for _, k := range ch.AllowedKeys {
					allow[k] = true
				}
				s.AllowedKeys = allow
			}
			for _, k := range keys {
				if ev := s.Press(k); ev.Kind == engine.EvInvalid {
					t.Errorf("%s: solution key %q rejected as invalid (not in AllowedKeys?)", ch.ID, k)
				}
			}
			if !ch.Goal.Met(s) {
				t.Errorf("%s: goal NOT met after optimal solution %v; buffer=%q cursor=%v",
					ch.ID, keys, s.Buffer, s.Cursor)
			}
			if len(keys) > ch.Par {
				t.Errorf("%s: solution is %d keys but par is %d (3 stars impossible)",
					ch.ID, len(keys), ch.Par)
			}
		}
	}

	// Guard against a stale solution map if challenges are added/removed later.
	for id := range solutions {
		if !seen[id] {
			t.Errorf("solution authored for %q but no such challenge exists", id)
		}
	}
}
