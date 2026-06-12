package progress

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveLoadAndReset(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "nested", "progress.json")}
	want := Model{
		XP:                  25,
		Streak:              2,
		CompletedChallenges: []string{"one", "two"},
		Badges:              []string{"First Steps"},
		LastLessonID:        "002-motions",
	}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if err := store.Reset(); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, Model{}) {
		t.Fatalf("after reset = %#v", got)
	}
}
