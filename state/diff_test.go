package state_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/state"
)

func TestDiffBetweenLinearChainLenAndApply(t *testing.T) {
	t.Parallel()

	// 0 → +1 → *10 → +2 == 12; diff from after first step covers *10 then +2.
	s0 := state.Of(0)
	s1 := s0.Apply(func(v int) int { return v + 1 })
	s2 := s1.Apply(func(v int) int { return v * 10 })
	s3 := s2.Apply(func(v int) int { return v + 2 })

	d := state.DiffBetween(s1, s3)
	if got := d.Len(); got != 2 {
		t.Fatalf("DiffBetween(s1, s3).Len() = %d, want 2", got)
	}
	if got := d.From(); got != s1.Version() {
		t.Fatalf("Diff.From() = %d, want %d", got, s1.Version())
	}
	if got := d.To(); got != s3.Version() {
		t.Fatalf("Diff.To() = %d, want %d", got, s3.Version())
	}

	applied := d.Apply(s1)
	if got := applied.Version(); got != s3.Version() {
		t.Fatalf("Diff.Apply(s1).Version() = %d, want %d", got, s3.Version())
	}
	if got := applied.Materialize(); got != s3.Materialize() {
		t.Fatalf("Diff.Apply(s1).Materialize() = %d, want %d", got, s3.Materialize())
	}
	// Apply is lazy: replaying the plan must not force materialization of s3.
	if got := s3.Materialize(); got != 12 {
		t.Fatalf("s3.Materialize() = %d, want 12", got)
	}
}

func TestDiffBetweenRootToTip(t *testing.T) {
	t.Parallel()

	root := state.Of(0)
	tip := root.
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 2 })

	d := state.DiffBetween(root, tip)
	if got := d.Len(); got != 3 {
		t.Fatalf("DiffBetween(root, tip).Len() = %d, want 3", got)
	}
	applied := d.Apply(root)
	if got := applied.Materialize(); got != 12 {
		t.Fatalf("Diff.Apply(root).Materialize() = %d, want 12", got)
	}
	if got := applied.Version(); got != tip.Version() {
		t.Fatalf("Diff.Apply(root).Version() = %d, want %d", got, tip.Version())
	}
}

func TestDiffBetweenEmptySameState(t *testing.T) {
	t.Parallel()

	s := state.Of(10).
		Apply(func(v int) int { return v + 5 }).
		Apply(func(v int) int { return v * 2 })

	d := state.DiffBetween(s, s)
	if got := d.Len(); got != 0 {
		t.Fatalf("DiffBetween(s, s).Len() = %d, want 0", got)
	}
	if got := d.From(); got != s.Version() {
		t.Fatalf("Diff.From() = %d, want %d", got, s.Version())
	}
	if got := d.To(); got != s.Version() {
		t.Fatalf("Diff.To() = %d, want %d", got, s.Version())
	}

	applied := d.Apply(s)
	if got := applied.Materialize(); got != s.Materialize() {
		t.Fatalf("empty Diff.Apply Materialize() = %d, want %d", got, s.Materialize())
	}
	if got := applied.Version(); got != s.Version() {
		t.Fatalf("empty Diff.Apply Version() = %d, want %d", got, s.Version())
	}
}

func TestDiffBetweenSiblingBranchesPanic(t *testing.T) {
	t.Parallel()

	parent := state.Of(0).Apply(func(v int) int { return v + 1 })
	left := parent.Apply(func(v int) int { return v * 10 })
	right := parent.Apply(func(v int) int { return v + 100 })

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("DiffBetween(siblings) did not panic")
		}
		const want = "state: unrelated lineages"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	_ = state.DiffBetween(left, right)
}

func TestDiffBetweenUnrelatedOfRootsPanic(t *testing.T) {
	t.Parallel()

	a := state.Of(1).Apply(func(v int) int { return v + 1 })
	b := state.Of(1).Apply(func(v int) int { return v + 1 })

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("DiffBetween(unrelated Of roots) did not panic")
		}
		const want = "state: unrelated lineages"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	_ = state.DiffBetween(a, b)
}

func TestDiffBetweenZeroZeroEmpty(t *testing.T) {
	t.Parallel()

	var from, to state.State[int]
	d := state.DiffBetween(from, to)
	if got := d.Len(); got != 0 {
		t.Fatalf("DiffBetween(zero, zero).Len() = %d, want 0", got)
	}
	if got := d.From(); got != 0 {
		t.Fatalf("Diff.From() = %d, want 0", got)
	}
	if got := d.To(); got != 0 {
		t.Fatalf("Diff.To() = %d, want 0", got)
	}

	applied := d.Apply(from)
	if got := applied.Materialize(); got != 0 {
		t.Fatalf("empty zero Diff.Apply Materialize() = %d, want 0", got)
	}
	if got := applied.Version(); got != 0 {
		t.Fatalf("empty zero Diff.Apply Version() = %d, want 0", got)
	}
}

func TestDiffBetweenZeroAndOfPanic(t *testing.T) {
	t.Parallel()

	var zero state.State[int]
	of := state.Of(0)

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("DiffBetween(zero, Of) did not panic")
		}
		const want = "state: unrelated lineages"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	_ = state.DiffBetween(zero, of)
}

func TestDiffApplyDoesNotMaterialize(t *testing.T) {
	t.Parallel()

	var calls int
	inc := func(v int) int {
		calls++
		return v + 1
	}

	from := state.Of(0)
	to := from.Apply(inc).Apply(inc)
	d := state.DiffBetween(from, to)
	applied := d.Apply(from)

	if calls != 0 {
		t.Fatalf("calls after Diff.Apply = %d, want 0 (lazy)", calls)
	}
	if got := applied.Materialize(); got != 2 {
		t.Fatalf("applied.Materialize() = %d, want 2", got)
	}
	if calls != 2 {
		t.Fatalf("calls after Materialize = %d, want 2", calls)
	}
}

func TestDiffApplyWrongBasePanic(t *testing.T) {
	t.Parallel()

	from := state.Of(0)
	to := from.Apply(func(v int) int { return v + 1 })
	d := state.DiffBetween(from, to)
	// Same lineage but advanced past From — not plan-equivalent to from.
	wrong := to

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("Diff.Apply(wrong base) did not panic")
		}
		const want = "state: unrelated lineages"
		if got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	_ = d.Apply(wrong)
}

func TestDiffBetweenTransactionCommitLineage(t *testing.T) {
	t.Parallel()

	parent := state.Of(0).Apply(func(v int) int { return v + 1 })
	committed := parent.Begin().
		Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 2 }).
		Commit()

	d := state.DiffBetween(parent, committed)
	if got := d.Len(); got != 2 {
		t.Fatalf("DiffBetween(parent, committed).Len() = %d, want 2", got)
	}
	applied := d.Apply(parent)
	if got := applied.Materialize(); got != committed.Materialize() {
		t.Fatalf("Diff.Apply Materialize() = %d, want %d", got, committed.Materialize())
	}
}
