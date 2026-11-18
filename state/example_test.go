package state_test

import (
	"fmt"

	"go.alvadorn.org/functional/state"
)

// ExampleOf shows the core loop: Of seeds a Version-0 plan, Apply records
// pure mutations without running them, and Materialize re-applies the full
// plan oldest→newest (no cache).
func ExampleOf() {
	s0 := state.Of(0)
	s1 := s0.Apply(func(v int) int { return v + 1 })
	s2 := s1.Apply(func(v int) int { return v * 10 })

	fmt.Println(s0.Version(), s0.Materialize())
	fmt.Println(s1.Version(), s1.Materialize())
	fmt.Println(s2.Version(), s2.Materialize())
	// Output:
	// 0 0
	// 1 1
	// 2 10
}

// ExampleTransaction shows Begin/Apply/Commit batching plan steps without
// materializing. Commit returns a State; Materialize is a separate call.
// Transaction is single-use: further Apply/Commit after Commit panics.
func ExampleTransaction() {
	parent := state.Of(1).Apply(func(v int) int { return v + 1 }) // → 2

	committed := parent.Begin().
		Apply(func(v int) int { return v * 10 }).
		Apply(func(v int) int { return v + 3 }).
		Commit()

	fmt.Println(parent.Version(), parent.Materialize())
	fmt.Println(committed.Version(), committed.Materialize())
	// Output:
	// 1 2
	// 3 23
}

// ExampleSnapshot freezes Version and Value at one materialization point.
// Further Apply on the source State does not change an already-taken Snapshot.
func ExampleSnapshot() {
	s := state.Of(5).Apply(func(v int) int { return v * 2 }) // → 10, Version 1
	snap := s.Snapshot()

	later := s.Apply(func(v int) int { return v + 1 }) // → 11, Version 2

	fmt.Println(snap.Version, snap.Value)
	fmt.Println(later.Version(), later.Materialize())
	// Output:
	// 1 10
	// 2 11
}
