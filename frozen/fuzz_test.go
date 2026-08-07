package frozen_test

import (
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

// FuzzSliceRoundTrip treats each fuzzed byte as one int element and checks
// the invariants NewSlice/Items/Values must uphold for any content: Len
// matches, Items reconstructs the exact content, and mutating one Items()
// result never reaches a later call.
func FuzzSliceRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{1, 2, 3})
	f.Add([]byte("hello, frozen"))

	f.Fuzz(func(t *testing.T, data []byte) {
		want := make([]int, len(data))
		for i, b := range data {
			want[i] = int(b)
		}

		s := frozen.NewSlice(want...)

		if got := s.Len(); got != len(want) {
			t.Fatalf("Len() = %d, want %d", got, len(want))
		}

		first := s.Items()
		if !slices.Equal(first, want) {
			t.Fatalf("Items() = %v, want %v", first, want)
		}

		if len(first) > 0 {
			first[0] = -1
			second := s.Items()
			if !slices.Equal(second, want) {
				t.Fatalf("mutating an Items() result leaked into a later call: got %v, want %v", second, want)
			}
		}

		i := 0
		for v := range s.Values() {
			if v != want[i] {
				t.Fatalf("Values() at %d = %v, want %v", i, v, want[i])
			}
			i++
		}
		if i != len(want) {
			t.Fatalf("Values() produced %d elements, want %d", i, len(want))
		}
	})
}
