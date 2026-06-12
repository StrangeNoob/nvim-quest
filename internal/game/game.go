// Package game holds the scoring and progression math.
package game

const MaxCombo = 5

// Stars rates a clear: at or under par 3, within par+2 2, otherwise 1.
func Stars(keystrokes, par int) int {
	switch {
	case keystrokes <= par:
		return 3
	case keystrokes <= par+2:
		return 2
	default:
		return 1
	}
}

// NextCombo raises the multiplier after a clean clear, capped at MaxCombo.
func NextCombo(c int) int {
	if c < MaxCombo {
		return c + 1
	}
	return MaxCombo
}

// XP is the base award scaled by the active combo multiplier.
func XP(base, combo int) int { return base * combo }

// LevelForXP: reaching level N+1 costs 100*N XP beyond level N
// (thresholds 0, 100, 300, 600, 1000, ...).
func LevelForXP(xp int) int {
	level, need := 1, 0
	for {
		need += 100 * level
		if xp < need {
			return level
		}
		level++
	}
}

// Badge names.
const (
	BadgeFirstSteps  = "First Steps"
	BadgeBirdie      = "Birdie"
	BadgeUntouchable = "Untouchable"
	BadgeGridBreaker = "Grid Breaker"
)

// ActBadge names the badge for completing an act's boss.
func ActBadge(act int) string {
	switch act {
	case 1:
		return "Dojo Graduate"
	case 2:
		return "Crypt Conqueror"
	case 3:
		return "Grid Runner"
	}
	return ""
}

// BirdieEarned reports whether 10+ challenges have been cleared at 3 stars.
func BirdieEarned(stars map[string]int) bool {
	count := 0
	for _, s := range stars {
		if s == 3 {
			count++
		}
	}
	return count >= 10
}
