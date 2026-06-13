package update

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDue(t *testing.T) {
	now := time.Now()
	if Due(Cache{LastCheck: now.Add(-12 * time.Hour)}, now, CheckInterval) {
		t.Error("12h since last check should not be due (interval 24h)")
	}
	if !Due(Cache{LastCheck: now.Add(-25 * time.Hour)}, now, CheckInterval) {
		t.Error("25h since last check should be due")
	}
	if !Due(Cache{}, now, CheckInterval) {
		t.Error("zero-value cache (never checked) should be due")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "update-check.json")
	want := Cache{LastCheck: time.Now().Truncate(time.Second), LatestVersion: "v0.2.0"}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got := Load(path)
	if !got.LastCheck.Equal(want.LastCheck) || got.LatestVersion != want.LatestVersion {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func TestLoadMissingOrCorruptIsZero(t *testing.T) {
	if c := Load(filepath.Join(t.TempDir(), "nope.json")); c != (Cache{}) {
		t.Errorf("missing file = %+v, want zero", c)
	}
}

func TestShouldCheck(t *testing.T) {
	tests := []struct {
		name    string
		noFlag  bool
		env     string
		version string
		want    bool
	}{
		{"normal versioned build", false, "", "v0.1.0", true},
		{"flag disables", true, "", "v0.1.0", false},
		{"env disables", false, "1", "v0.1.0", false},
		{"dev build skips", false, "", "dev", false},
	}
	for _, tt := range tests {
		if got := ShouldCheck(tt.noFlag, tt.env, tt.version); got != tt.want {
			t.Errorf("%s: ShouldCheck(%v,%q,%q) = %v, want %v",
				tt.name, tt.noFlag, tt.env, tt.version, got, tt.want)
		}
	}
}
