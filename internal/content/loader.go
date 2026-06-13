package content

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"nvim-quest/assets"
)

// All returns every embedded lesson, sorted by (act, order).
func All() ([]Lesson, error) {
	entries, err := assets.Lessons.ReadDir("lessons")
	if err != nil {
		return nil, fmt.Errorf("read embedded lessons: %w", err)
	}
	var lessons []Lesson
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := assets.Lessons.ReadFile("lessons/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		var l Lesson
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		lessons = append(lessons, l)
	}
	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].Act != lessons[j].Act {
			return lessons[i].Act < lessons[j].Act
		}
		return lessons[i].Order < lessons[j].Order
	})
	return lessons, nil
}
