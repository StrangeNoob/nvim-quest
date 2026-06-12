package progress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "progress.json")
	p := New()
	p.XP = 340
	p.Level = 2
	p.Stars["a1l1c1"] = 3
	p.MarkCompleted("a1l1c1")
	p.AddBadge("First Steps")
	p.LastLesson = "act1-01-the-two-stances"
	if err := Save(path, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got := Load(path)
	if got.XP != 340 || got.Level != 2 || got.Stars["a1l1c1"] != 3 ||
		!got.IsCompleted("a1l1c1") || !got.HasBadge("First Steps") ||
		got.LastLesson != "act1-01-the-two-stances" {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestLoadMissingFileStartsFresh(t *testing.T) {
	p := Load(filepath.Join(t.TempDir(), "nope.json"))
	if p.Version != 2 || p.XP != 0 || p.Level != 1 || p.Stars == nil {
		t.Errorf("fresh progress malformed: %+v", p)
	}
}

func TestLoadV1FileBacksUpAndStartsFresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	v1 := []byte(`{"xp":120,"streak":3,"completed_challenges":["c1"]}`)
	if err := os.WriteFile(path, v1, 0o644); err != nil {
		t.Fatal(err)
	}
	p := Load(path)
	if p.XP != 0 {
		t.Errorf("v1 file must not load as v2, got XP %d", p.XP)
	}
	if _, err := os.Stat(path + ".v1.bak"); err != nil {
		t.Errorf("expected backup file: %v", err)
	}
}

func TestLoadCorruptFileStartsFresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	os.WriteFile(path, []byte("{not json"), 0o644)
	p := Load(path)
	if p.Version != 2 {
		t.Errorf("corrupt file must yield fresh progress: %+v", p)
	}
}

func TestDoubleCompleteIsIdempotent(t *testing.T) {
	p := New()
	p.MarkCompleted("x")
	p.MarkCompleted("x")
	if len(p.Completed) != 1 {
		t.Errorf("Completed = %v, want single entry", p.Completed)
	}
	p.AddBadge("B")
	p.AddBadge("B")
	if len(p.Badges) != 1 {
		t.Errorf("Badges = %v, want single entry", p.Badges)
	}
}
