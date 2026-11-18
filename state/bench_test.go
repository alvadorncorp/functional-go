package state_test

import (
	"fmt"
	"testing"

	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/state"
)

// Plan sizes for marginal Apply cost. A true O(1) reverse-linked prepend
// keeps allocs/op ~constant (~1 plan node) across these n values.
var applyBenchSizes = []int{10, 100, 1_000, 10_000}

// Materialize plan lengths. Full plan walk + temporary steps slice scale
// with n; ReportAllocs surfaces that cost separately from Apply.
var materializeBenchSizes = []int{1, 10, 100, 1_000}

// Prevent the compiler from eliding Apply's plan-node allocation.
var applySink state.State[int]

// Prevent the compiler from eliding Materialize / Commit results.
var (
	materializeSink int
	commitSink      state.State[int]
	immutableSink   state.State[model]
)

// BenchmarkStateApply measures ONE additional Apply on a plan already of
// size n. Setup builds the n-step plan outside the timed region so the
// reported ns/op and allocs/op are the marginal cost of plan growth — not
// the cost of building the whole chain. With reverse-linked prepend,
// allocs/op must stay ~constant in n (one planNode per Apply).
func BenchmarkStateApply(b *testing.B) {
	inc := func(v int) int { return v + 1 }

	for _, n := range applyBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := state.Of(0)
			for range n {
				s = s.Apply(inc)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				applySink = s.Apply(inc)
			}
		})
	}
}

// BenchmarkStateMaterialize measures one Materialize of a plan of size n.
// Setup builds the plan outside the timed region so ns/op and allocs/op
// reflect the re-apply walk (and temporary steps slice), not plan growth.
func BenchmarkStateMaterialize(b *testing.B) {
	inc := func(v int) int { return v + 1 }

	for _, n := range materializeBenchSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			s := state.Of(0)
			for range n {
				s = s.Apply(inc)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				materializeSink = s.Materialize()
			}
		})
	}
}

// BenchmarkTransactionApplyCommit compares batching k mutations via a
// single-use Transaction against the same k mutations applied as sequential
// State.Apply notes. Setup is minimal; the timed path includes every Apply
// plus Commit (or the final note State).
func BenchmarkTransactionApplyCommit(b *testing.B) {
	const batch = 16
	inc := func(v int) int { return v + 1 }
	base := state.Of(0)

	b.Run("tx", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			tx := base.Begin()
			for range batch {
				tx.Apply(inc)
			}
			commitSink = tx.Commit()
		}
	})

	b.Run("note", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s := base
			for range batch {
				s = s.Apply(inc)
			}
			commitSink = s
		}
	})
}

// model is a cheap T whose field is a immutable.Map. Production state does
// not import immutable; this bench lives in state_test only.
type model struct {
	Attrs immutable.Map[string, int]
}

// BenchmarkImmutableInState measures Apply + Materialize when T embeds a
// immutable.Map field. Each mutation path-copies one Map entry; the State
// plan itself stays reverse-linked and mesh-independent.
func BenchmarkImmutableInState(b *testing.B) {
	root := state.Of(model{Attrs: immutable.MapOf[string, int](nil)})
	set := func(m model) model {
		m.Attrs = m.Attrs.Set("k", m.Attrs.Len()+1)
		return m
	}
	// Pre-grow a small plan so Materialize does non-trivial work.
	s := root
	for range 8 {
		s = s.Apply(set)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		next := s.Apply(set)
		_ = next.Materialize()
		immutableSink = next
	}
}
