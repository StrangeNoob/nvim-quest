package update

import "testing"

// TestNewUpdaterConstructs exercises the updater construction (checksum
// validator wiring) without touching the network. Live release detection is
// intentionally not tested here — it would require a published release and a
// network call, making the suite flaky.
func TestNewUpdaterConstructs(t *testing.T) {
	up, err := newUpdater()
	if err != nil {
		t.Fatalf("newUpdater: %v", err)
	}
	if up == nil {
		t.Fatal("newUpdater returned nil updater")
	}
}
