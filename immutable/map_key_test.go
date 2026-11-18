package immutable_test

import (
	"math"
	"testing"

	"go.alvadorn.org/functional/immutable"
)

func assertStoredZero(t *testing.T, m immutable.Map[float64, string], negative bool, value string) {
	t.Helper()
	zeros := 0
	for k, v := range m.All() {
		if k == 0 {
			zeros++
			if math.Signbit(k) != negative || v != value {
				t.Errorf("stored zero = (negative=%v, %q), want (%v, %q)", math.Signbit(k), v, negative, value)
			}
		}
	}
	if zeros != 1 {
		t.Errorf("stored zero keys = %d, want 1", zeros)
	}
	if got, ok := m.Get(0); !ok || got != value {
		t.Errorf("Get(0) = (%q, %v), want (%q, true)", got, ok, value)
	}
}

func TestMapLastEquivalentKeyWins(t *testing.T) {
	t.Parallel()

	for _, negative := range []bool{false, true} {
		first, last := math.Copysign(0, -1), 0.0
		if negative {
			first, last = last, first
		}
		base := immutable.MapOf(map[float64]string{first: "first"})
		assertStoredZero(t, base.Set(last, "last"), negative, "last")

		b := base.Builder()
		b.Set(last, "last") // Copy a published leaf before replacing its key.
		assertStoredZero(t, b.Build(), negative, "last")
		b = immutable.MapOf[float64, string](nil).Builder()
		b.Set(first, "first").Set(last, "last") // Replace an already-owned leaf.
		assertStoredZero(t, b.Build(), negative, "last")

		collected := immutable.CollectMap(func(yield func(float64, string) bool) {
			if !yield(first, "first") {
				return
			}
			yield(last, "last")
		})
		assertStoredZero(t, collected, negative, "last")
		if collected.Len() != 1 {
			t.Errorf("CollectMap duplicate zero Len = %d, want 1", collected.Len())
		}

		// Two entries exercise MapOf's bulk path; the source already retains last.
		source := map[float64]string{first: "first", 1: "other"}
		source[last] = "last"
		assertStoredZero(t, immutable.MapOf(source), negative, "last")
		assertStoredZero(t, immutable.MapOf(map[float64]string{last: "last"}), negative, "last")
		assertStoredZero(t, base, !negative, "first")
	}
}
