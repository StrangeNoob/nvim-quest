package lessons

import "testing"

func TestLoadBundledLessons(t *testing.T) {
	all, err := NewLoader("../../lessons").All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 7 {
		t.Fatalf("loaded %d lessons, want at least 7", len(all))
	}
	for _, lesson := range all {
		if len(lesson.Challenges) == 0 {
			t.Fatalf("lesson %s has no challenges", lesson.ID)
		}
	}
}
