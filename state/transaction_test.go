package state_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/state"
)

func TestTransactionMultipleMutationsCommitMaterializes(t *testing.T) {
	t.Parallel()

	// Parent has one planned step: 0 → +1. Tx adds *10 then +2 → 12.
	parent := state.Of(0).Apply(func(v int) int { return v + 1 })
	tx := parent.Begin()
	tx.Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 2 })
	committed := tx.Commit()

	if got := committed.Materialize(); got != 12 {
		t.Fatalf("committed.Materialize() = %d, want 12", got)
	}
	if got := committed.Version(); got != parent.Version()+2 {
		t.Fatalf("committed.Version() = %d, want %d", got, parent.Version()+2)
	}
}

func TestTransactionParentRemainsUnchanged(t *testing.T) {
	t.Parallel()

	parent := state.Of(10).Apply(func(v int) int { return v + 5 }) // → 15, Version 1
	wantMat := parent.Materialize()
	wantVer := parent.Version()

	tx := parent.Begin()
	tx.Apply(func(v int) int { return v * 100 })
	committed := tx.Commit()

	if got := parent.Materialize(); got != wantMat {
		t.Fatalf("parent.Materialize() after Commit = %d, want %d", got, wantMat)
	}
	if got := parent.Version(); got != wantVer {
		t.Fatalf("parent.Version() after Commit = %d, want %d", got, wantVer)
	}
	if got := committed.Materialize(); got != 1500 {
		t.Fatalf("committed.Materialize() = %d, want 1500", got)
	}
	if got := committed.Version(); got != wantVer+1 {
		t.Fatalf("committed.Version() = %d, want %d", got, wantVer+1)
	}
}

func TestTransactionEmptyCommitEquivalentToSnapshot(t *testing.T) {
	t.Parallel()

	parent := state.Of(7).
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v * 3 }) // → 24, Version 2

	committed := parent.Begin().Commit()

	if got := committed.Materialize(); got != parent.Materialize() {
		t.Fatalf("empty Commit Materialize() = %d, want %d", got, parent.Materialize())
	}
	if got := committed.Version(); got != parent.Version() {
		t.Fatalf("empty Commit Version() = %d, want %d", got, parent.Version())
	}
}

func TestTransactionUseAfterCommitPanicsApply(t *testing.T) {
	t.Parallel()

	tx := state.Of(0).Begin()
	_ = tx.Commit()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("Apply after Commit did not panic")
		}
		const want = "state: transaction used after Commit"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	tx.Apply(func(v int) int { return v + 1 })
}

func TestTransactionUseAfterCommitPanicsCommit(t *testing.T) {
	t.Parallel()

	tx := state.Of(0).Begin()
	_ = tx.Commit()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("second Commit did not panic")
		}
		const want = "state: transaction used after Commit"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	_ = tx.Commit()
}

func TestTransactionCommitDoesNotRunMutations(t *testing.T) {
	t.Parallel()

	var calls int
	inc := func(v int) int {
		calls++
		return v + 1
	}

	parent := state.Of(0).Apply(inc) // planned, not run
	tx := parent.Begin()
	tx.Apply(inc).Apply(inc)
	committed := tx.Commit()

	if calls != 0 {
		t.Fatalf("calls after Commit = %d, want 0 (lazy)", calls)
	}

	if got := committed.Materialize(); got != 3 {
		t.Fatalf("committed.Materialize() = %d, want 3", got)
	}
	if calls != 3 {
		t.Fatalf("calls after Materialize = %d, want 3", calls)
	}
}
