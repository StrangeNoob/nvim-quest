package update

import "testing"

func TestNewer(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{"v0.1.0", "v0.2.0", true},
		{"0.1.0", "0.2.0", true},   // missing v prefix is canonicalized
		{"v0.1.5", "0.1.5", false}, // equal (mixed prefix)
		{"v0.2.0", "v0.1.0", false},
		{"v0.2.0", "v0.2.0", false},
		{"dev", "v1.0.0", false},     // dev is never "older"
		{"v0.1.0", "garbage", false}, // unparseable latest
	}
	for _, tt := range tests {
		if got := Newer(tt.current, tt.latest); got != tt.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestUpToDate(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{"v0.2.0", "v0.1.0", true},
		{"v0.2.0", "v0.2.0", true},
		{"v0.1.0", "v0.2.0", false},
		{"dev", "v1.0.0", false},     // dev should always upgrade on explicit update
		{"v1.0.0", "garbage", true},  // valid current, unparseable latest: treat as up to date
		{"garbage", "v1.0.0", false}, // unparseable current: not up to date
	}
	for _, tt := range tests {
		if got := upToDate(tt.current, tt.latest); got != tt.want {
			t.Errorf("upToDate(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}
