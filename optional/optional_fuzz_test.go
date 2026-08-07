package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

// step is the deterministic function used to check the monad laws below.
func step(x int) optional.Optional[int] {
	return optional.Some(x + 1)
}

// FuzzMonadLaws checks the three monad laws for Optional[int] and FlatMap:
// left identity, right identity, and associativity.
func FuzzMonadLaws(f *testing.F) {
	f.Add(0, true)
	f.Add(1, false)
	f.Add(-5, true)
	f.Add(100, false)

	f.Fuzz(func(t *testing.T, x int, present bool) {
		o := optional.None[int]()
		if present {
			o = optional.Some(x)
		}

		// Left identity: FlatMap(Some(x), step) == step(x)
		if left, want := optional.FlatMap(optional.Some(x), step), step(x); !left.Equal(want) {
			t.Fatalf("left identity: FlatMap(Some(%d), step) = %v, want %v", x, left, want)
		}

		// Right identity: FlatMap(o, Some) == o
		if right := optional.FlatMap(o, optional.Some[int]); !right.Equal(o) {
			t.Fatalf("right identity: FlatMap(%v, Some) = %v, want %v", o, right, o)
		}

		// Associativity:
		// FlatMap(FlatMap(o, step), step) == FlatMap(o, func(x){ return FlatMap(step(x), step) })
		assocLeft := optional.FlatMap(optional.FlatMap(o, step), step)
		assocRight := optional.FlatMap(o, func(v int) optional.Optional[int] {
			return optional.FlatMap(step(v), step)
		})
		if !assocLeft.Equal(assocRight) {
			t.Fatalf("associativity: got %v, want %v", assocLeft, assocRight)
		}
	})
}
