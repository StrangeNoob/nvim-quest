// Package progress persists the player's journey to a versioned JSON file.
package progress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

const Version = 2

type Progress struct {
	Version    int            `json:"version"`
	XP         int            `json:"xp"`
	Level      int            `json:"level"`
	Stars      map[string]int `json:"stars"`
	Completed  []string       `json:"completed"`
	Badges     []string       `json:"badges"`
	LastLesson string         `json:"lastLesson"`
}

func New() *Progress {
	return &Progress{Version: Version, Level: 1, Stars: map[string]int{}}
}

func (p *Progress) IsCompleted(id string) bool { return slices.Contains(p.Completed, id) }

func (p *Progress) MarkCompleted(id string) {
	if !p.IsCompleted(id) {
		p.Completed = append(p.Completed, id)
	}
}

func (p *Progress) HasBadge(name string) bool { return slices.Contains(p.Badges, name) }

func (p *Progress) AddBadge(name string) {
	if !p.HasBadge(name) {
		p.Badges = append(p.Badges, name)
	}
}

// DefaultPath is ~/.nvim-quest/progress.json.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "progress.json"
	}
	return filepath.Join(home, ".nvim-quest", "progress.json")
}

// Load never fails: missing, corrupt, or old-version files yield a fresh
// Progress (old files are kept as <path>.v1.bak).
func Load(path string) *Progress {
	data, err := os.ReadFile(path)
	if err != nil {
		return New()
	}
	var p Progress
	if err := json.Unmarshal(data, &p); err != nil || p.Version != Version {
		_ = os.Rename(path, path+".v1.bak")
		return New()
	}
	if p.Stars == nil {
		p.Stars = map[string]int{}
	}
	if p.Level < 1 {
		p.Level = 1
	}
	return &p
}

func Save(path string, p *Progress) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
