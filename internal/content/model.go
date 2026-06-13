// Package content loads the embedded lesson JSON files.
package content

import "github.com/StrangeNoob/nvim-quest/internal/engine"

type Lesson struct {
	ID         string      `json:"id"`
	Act        int         `json:"act"`
	Order      int         `json:"order"`
	Title      string      `json:"title"`
	Story      string      `json:"story"`
	Challenges []Challenge `json:"challenges"`
	Boss       *Boss       `json:"boss,omitempty"`
}

type Challenge struct {
	ID          string      `json:"id"`
	Intro       string      `json:"intro"`
	Buffer      []string    `json:"buffer"`
	Cursor      [2]int      `json:"cursor"`
	Goal        engine.Goal `json:"goal"`
	Par         int         `json:"par"`
	XP          int         `json:"xp"`
	Hint        string      `json:"hint"`
	NewKeys     []string    `json:"newKeys,omitempty"`
	AllowedKeys []string    `json:"allowedKeys,omitempty"`
}

type Boss struct {
	Name         string      `json:"name"`
	Taunt        string      `json:"taunt"`
	TimeLimitSec int         `json:"timeLimitSec"`
	XP           int         `json:"xp"`
	Steps        []Challenge `json:"steps"`
}

func goalTypes() []string { return engine.GoalTypes }
