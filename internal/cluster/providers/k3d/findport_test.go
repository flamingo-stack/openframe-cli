package k3d

import "testing"

// TestFindPort_SkipsUsedPorts guards the property that matters for correctness:
// findPort never returns a port already taken by another cluster. A regression
// here (e.g. the used-ports map coming back empty) causes port conflicts on the
// next cluster create.
func TestFindPort_SkipsUsedPorts(t *testing.T) {
	m := &K3dManager{}
	used := map[int]bool{6550: true, 6551: true, 6552: true}

	// Preferred ports are all used → must fall through to the search range and
	// return a free port that is NOT in the used set.
	got := m.findPort([]int{6550, 6551, 6552}, 6553, used)

	if got == 0 {
		t.Fatal("expected a free port, got 0")
	}
	if used[got] {
		t.Fatalf("findPort returned a used port: %d", got)
	}
}

func TestFindPort_ExhaustedReturnsZero(t *testing.T) {
	m := &K3dManager{}
	// Mark the entire search window (searchStart .. searchStart+1000) plus the
	// preferred port as used, so no candidate is free → 0.
	used := map[int]bool{}
	const start = 20000
	for p := start; p < start+1000; p++ {
		used[p] = true
	}
	used[19999] = true

	if got := m.findPort([]int{19999}, start, used); got != 0 {
		t.Fatalf("expected 0 when every candidate is used, got %d", got)
	}
}
