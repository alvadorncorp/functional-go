package state_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/state"
)

func TestMaterializeZeroMutationsReturnsBase(t *testing.T) {
	t.Parallel()

	const base = 42
	s := state.Of(base)

	if got := s.Materialize(); got != base {
		t.Fatalf("Materialize() = %d, want %d", got, base)
	}
}

func TestMaterializeOneMutation(t *testing.T) {
	t.Parallel()

	s := state.Of(10).Apply(func(v int) int { return v + 5 })

	if got := s.Materialize(); got != 15 {
		t.Fatalf("Materialize() = %d, want 15", got)
	}
}

func TestMaterializeMultipleMutationsOrder(t *testing.T) {
	t.Parallel()

	// Plan A then B then C: start 0 → +1 → *10 → +2 == 12.
	s := state.Of(0).
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 2 })

	if got := s.Materialize(); got != 12 {
		t.Fatalf("Materialize() = %d, want 12 (order A then B then C)", got)
	}
}

func TestApplyBranchingLeavesParentUnchanged(t *testing.T) {
	t.Parallel()

	s1 := state.Of(0).Apply(func(v int) int { return v + 1 }) // A
	s2 := s1.Apply(func(v int) int { return v * 10 })         // A then B
	s3 := s1.Apply(func(v int) int { return v + 100 })        // A then C

	if got := s1.Materialize(); got != 1 {
		t.Fatalf("s1.Materialize() = %d, want 1 (parent unchanged)", got)
	}
	if got := s2.Materialize(); got != 10 {
		t.Fatalf("s2.Materialize() = %d, want 10", got)
	}
	if got := s3.Materialize(); got != 101 {
		t.Fatalf("s3.Materialize() = %d, want 101", got)
	}
}

func TestApplyDoesNotRunMutation(t *testing.T) {
	t.Parallel()

	var calls int
	inc := func(v int) int {
		calls++
		return v + 1
	}

	s := state.Of(0).Apply(inc).Apply(inc)
	if calls != 0 {
		t.Fatalf("calls after Apply = %d, want 0", calls)
	}

	_ = s.Materialize()
	if calls != 2 {
		t.Fatalf("calls after first Materialize = %d, want 2", calls)
	}
}

func TestMaterializeReappliesEveryCall(t *testing.T) {
	t.Parallel()

	var calls int
	inc := func(v int) int {
		calls++
		return v + 1
	}

	s := state.Of(0).Apply(inc)

	if got := s.Materialize(); got != 1 {
		t.Fatalf("first Materialize() = %d, want 1", got)
	}
	if calls != 1 {
		t.Fatalf("calls after first Materialize = %d, want 1", calls)
	}

	if got := s.Materialize(); got != 1 {
		t.Fatalf("second Materialize() = %d, want 1", got)
	}
	if calls != 2 {
		t.Fatalf("calls after second Materialize = %d, want 2 (re-applied)", calls)
	}
}

func TestZeroValueStateMaterialize(t *testing.T) {
	t.Parallel()

	var s state.State[int]
	if got := s.Materialize(); got != 0 {
		t.Fatalf("zero State.Materialize() = %d, want 0", got)
	}
}

func TestApplyBranchingSharedPrefixIndependence(t *testing.T) {
	t.Parallel()

	// Common prefix of 100 +1 steps → base materializes to 100.
	base := state.Of(0)
	for range 100 {
		base = base.Apply(func(v int) int { return v + 1 })
	}

	// left: +1 fifty more times → 150.
	left := base
	for range 50 {
		left = left.Apply(func(v int) int { return v + 1 })
	}

	// right: +10 thirty times → 400. Shares the 100-step prefix with left.
	right := base
	for range 30 {
		right = right.Apply(func(v int) int { return v + 10 })
	}

	if got := base.Materialize(); got != 100 {
		t.Fatalf("base.Materialize() = %d, want 100 (shared prefix unchanged)", got)
	}
	if got := left.Materialize(); got != 150 {
		t.Fatalf("left.Materialize() = %d, want 150", got)
	}
	if got := right.Materialize(); got != 400 {
		t.Fatalf("right.Materialize() = %d, want 400", got)
	}

	// Further evolution on left must not affect right or base.
	left2 := left.Apply(func(v int) int { return v * 2 })
	if got := left2.Materialize(); got != 300 {
		t.Fatalf("left2.Materialize() = %d, want 300", got)
	}
	if got := left.Materialize(); got != 150 {
		t.Fatalf("left.Materialize() after branch = %d, want 150", got)
	}
	if got := right.Materialize(); got != 400 {
		t.Fatalf("right.Materialize() after left branch = %d, want 400", got)
	}
	if got := base.Materialize(); got != 100 {
		t.Fatalf("base.Materialize() after left branch = %d, want 100", got)
	}
}

func TestMaterializeLargeChain(t *testing.T) {
	t.Parallel()

	const n = 10_000
	s := state.Of(0)
	for range n {
		s = s.Apply(func(v int) int { return v + 1 })
	}

	if got := s.Materialize(); got != n {
		t.Fatalf("Materialize() after %d applies = %d, want %d", n, got, n)
	}
}

func TestOfStartsAtVersionZero(t *testing.T) {
	t.Parallel()

	s := state.Of(42)
	if got := s.Version(); got != 0 {
		t.Fatalf("Of(...).Version() = %d, want 0", got)
	}
}

func TestZeroValueStateVersionZero(t *testing.T) {
	t.Parallel()

	var s state.State[int]
	if got := s.Version(); got != 0 {
		t.Fatalf("zero State.Version() = %d, want 0", got)
	}
}

func TestApplyIncrementsVersionByOne(t *testing.T) {
	t.Parallel()

	s0 := state.Of(0)
	s1 := s0.Apply(func(v int) int { return v + 1 })
	s2 := s1.Apply(func(v int) int { return v + 1 })
	s3 := s2.Apply(func(v int) int { return v + 1 })

	if got := s0.Version(); got != 0 {
		t.Fatalf("s0.Version() = %d, want 0", got)
	}
	if got := s1.Version(); got != 1 {
		t.Fatalf("s1.Version() = %d, want 1", got)
	}
	if got := s2.Version(); got != 2 {
		t.Fatalf("s2.Version() = %d, want 2", got)
	}
	if got := s3.Version(); got != 3 {
		t.Fatalf("s3.Version() = %d, want 3", got)
	}
}

func TestApplyLeavesParentVersionUnchanged(t *testing.T) {
	t.Parallel()

	parent := state.Of(0).Apply(func(v int) int { return v + 1 }) // Version 1
	wantParent := parent.Version()
	_ = parent.Apply(func(v int) int { return v + 100 })

	if got := parent.Version(); got != wantParent {
		t.Fatalf("parent.Version() after child Apply = %d, want %d", got, wantParent)
	}
}

func TestApplyBranchingVersionsIndependent(t *testing.T) {
	t.Parallel()

	parent := state.Of(0).Apply(func(v int) int { return v + 1 }) // A → Version 1
	left := parent.Apply(func(v int) int { return v * 10 })       // A then B
	right := parent.Apply(func(v int) int { return v + 100 })     // A then C

	wantChild := parent.Version() + 1
	if got := left.Version(); got != wantChild {
		t.Fatalf("left.Version() = %d, want %d", got, wantChild)
	}
	if got := right.Version(); got != wantChild {
		t.Fatalf("right.Version() = %d, want %d", got, wantChild)
	}
	if got := parent.Version(); got != 1 {
		t.Fatalf("parent.Version() after branches = %d, want 1", got)
	}

	if got := parent.Materialize(); got != 1 {
		t.Fatalf("parent.Materialize() = %d, want 1", got)
	}
	if got := left.Materialize(); got != 10 {
		t.Fatalf("left.Materialize() = %d, want 10", got)
	}
	if got := right.Materialize(); got != 101 {
		t.Fatalf("right.Materialize() = %d, want 101", got)
	}
}

func TestSnapshotValueMatchesMaterialize(t *testing.T) {
	t.Parallel()

	s := state.Of(0).
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 2 })

	snap := s.Snapshot()
	want := s.Materialize()
	if snap.Value != want {
		t.Fatalf("Snapshot().Value = %d, want Materialize() %d", snap.Value, want)
	}
}

func TestSnapshotVersionMatchesVersion(t *testing.T) {
	t.Parallel()

	s := state.Of(0).
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v + 1 })

	snap := s.Snapshot()
	want := s.Version()
	if snap.Version != want {
		t.Fatalf("Snapshot().Version = %d, want Version() %d", snap.Version, want)
	}
}

func TestSnapshotUnchangedAfterStateEvolves(t *testing.T) {
	t.Parallel()

	s := state.Of(10).Apply(func(v int) int { return v + 5 }) // Value 15, Version 1
	snap := s.Snapshot()
	wantValue := snap.Value
	wantVersion := snap.Version

	// Evolve beyond the snapshot point.
	evolved := s.Apply(func(v int) int { return v * 2 }) // Value 30, Version 2
	if got := evolved.Materialize(); got != 30 {
		t.Fatalf("evolved.Materialize() = %d, want 30", got)
	}
	if got := evolved.Version(); got != 2 {
		t.Fatalf("evolved.Version() = %d, want 2", got)
	}

	// Snapshot taken earlier must remain frozen.
	if snap.Value != wantValue {
		t.Fatalf("snap.Value after evolve = %d, want %d", snap.Value, wantValue)
	}
	if snap.Version != wantVersion {
		t.Fatalf("snap.Version after evolve = %d, want %d", snap.Version, wantVersion)
	}
	if snap.Value != 15 {
		t.Fatalf("snap.Value = %d, want 15", snap.Value)
	}
	if snap.Version != 1 {
		t.Fatalf("snap.Version = %d, want 1", snap.Version)
	}

	// Source state value semantics: Apply returns a new State; s is unchanged.
	if got := s.Materialize(); got != 15 {
		t.Fatalf("s.Materialize() after branch = %d, want 15", got)
	}
	if got := s.Version(); got != 1 {
		t.Fatalf("s.Version() after branch = %d, want 1", got)
	}
}
