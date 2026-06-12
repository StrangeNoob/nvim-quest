package lessons

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Loader struct {
	Dir string
}

func NewLoader(dir string) Loader {
	return Loader{Dir: dir}
}

func (l Loader) All() ([]Lesson, error) {
	paths, err := filepath.Glob(filepath.Join(l.Dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	result := make([]Lesson, 0, len(paths))
	for _, path := range paths {
		lesson, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		result = append(result, lesson)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no lessons found in %s", l.Dir)
	}
	return result, nil
}

func (l Loader) ByID(id string) (Lesson, error) {
	all, err := l.All()
	if err != nil {
		return Lesson{}, err
	}
	for _, lesson := range all {
		if lesson.ID == id {
			return lesson, nil
		}
	}
	return Lesson{}, fmt.Errorf("lesson %q not found", id)
}

func (l Loader) ByTopic(topic string) ([]Lesson, error) {
	all, err := l.All()
	if err != nil {
		return nil, err
	}
	var result []Lesson
	for _, lesson := range all {
		if lesson.Topic == topic {
			result = append(result, lesson)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("practice topic %q not found", topic)
	}
	return result, nil
}

func loadFile(path string) (Lesson, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lesson{}, err
	}
	var lesson Lesson
	if err := json.Unmarshal(data, &lesson); err != nil {
		return Lesson{}, fmt.Errorf("%s: %w", path, err)
	}
	return lesson, nil
}
