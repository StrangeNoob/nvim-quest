package game

import "testing"

func TestStars(t *testing.T) {
	tests := []struct{ keys, par, want int }{
		{3, 4, 3}, {4, 4, 3}, {5, 4, 2}, {6, 4, 2}, {7, 4, 1}, {20, 4, 1},
	}
	for _, tt := range tests {
		if got := Stars(tt.keys, tt.par); got != tt.want {
			t.Errorf("Stars(%d, %d) = %d, want %d", tt.keys, tt.par, got, tt.want)
		}
	}
}

func TestNextCombo(t *testing.T) {
	tests := []struct{ in, want int }{{1, 2}, {4, 5}, {5, 5}}
	for _, tt := range tests {
		if got := NextCombo(tt.in); got != tt.want {
			t.Errorf("NextCombo(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestXP(t *testing.T) {
	if got := XP(50, 3); got != 150 {
		t.Errorf("XP(50, 3) = %d, want 150", got)
	}
}

func TestLevelForXP(t *testing.T) {
	tests := []struct{ xp, want int }{
		{0, 1}, {99, 1}, {100, 2}, {299, 2}, {300, 3}, {599, 3}, {600, 4},
	}
	for _, tt := range tests {
		if got := LevelForXP(tt.xp); got != tt.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tt.xp, got, tt.want)
		}
	}
}

func TestBirdieEarned(t *testing.T) {
	stars := map[string]int{}
	for i := 0; i < 9; i++ {
		stars[string(rune('a'+i))] = 3
	}
	if BirdieEarned(stars) {
		t.Error("9 three-star clears must not earn Birdie")
	}
	stars["j"] = 3
	if !BirdieEarned(stars) {
		t.Error("10 three-star clears must earn Birdie")
	}
}

func TestActBadge(t *testing.T) {
	tests := []struct {
		act  int
		want string
	}{
		{1, "Dojo Graduate"}, {2, "Crypt Conqueror"}, {3, "Grid Runner"}, {4, ""},
	}
	for _, tt := range tests {
		if got := ActBadge(tt.act); got != tt.want {
			t.Errorf("ActBadge(%d) = %q, want %q", tt.act, got, tt.want)
		}
	}
}
