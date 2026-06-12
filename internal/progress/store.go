package progress

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Store struct {
	Path string
}

func DefaultStore() (Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Store{}, err
	}
	return Store{Path: filepath.Join(home, ".nvim-quest", "progress.json")}, nil
}

func (s Store) Load() (Model, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Model{}, nil
	}
	if err != nil {
		return Model{}, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return Model{}, err
	}
	return model, nil
}

func (s Store) Save(model Model) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, append(data, '\n'), 0o644)
}

func (s Store) Reset() error {
	err := os.Remove(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
