package immutable_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// numReaders is the number of concurrent reader goroutines each
// concurrency test starts against one shared base collection.
const numReaders = 8

// numDerivations is the number of new versions the writer goroutine derives
// from the shared base while readers are running.
const numDerivations = 200

// TestListConcurrentReadDuringDerive proves that reading a shared List
// (Values/At) is unaffected by a concurrent goroutine deriving new List
// versions (Append) from that same base: under -race, none of this may
// report a data race, and every reader must keep seeing the base's original
// contents throughout.
func TestListConcurrentReadDuringDerive(t *testing.T) {
	t.Parallel()

	base := immutable.ListOf(1, 2, 3, 4, 5)
	want := base.Items()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for range numReaders {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				for i := range base.Len() {
					if base.At(i) != want[i] {
						t.Errorf("At(%d) = %d during concurrent derive, want %d", i, base.At(i), want[i])
						return
					}
				}
				sum := 0
				for v := range base.Values() {
					sum += v
				}
				if sum != 15 {
					t.Errorf("sum over Values() = %d during concurrent derive, want 15", sum)
					return
				}
			}
		})
	}

	go func() {
		v := base
		for range numDerivations {
			v = v.Append(v.Len())
		}
	}()

	// Give the reader goroutines a real window to race against the writer
	// before signaling them to stop.
	for range numDerivations {
		_ = base.Append(0)
	}
	close(stop)
	wg.Wait()

	if got := base.Items(); len(got) != len(want) {
		t.Fatalf("base changed after concurrent derivation: got %v, want %v", got, want)
	}
}

// TestMapConcurrentReadDuringDerive is the Map analogue of
// TestListConcurrentReadDuringDerive, deriving with Set instead of Append.
func TestMapConcurrentReadDuringDerive(t *testing.T) {
	t.Parallel()

	base := immutable.MapOf(map[string]int{"a": 1, "b": 2, "c": 3})

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for range numReaders {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if got, ok := base.Get("a"); got != 1 || !ok {
					t.Errorf("Get(a) = (%d,%v) during concurrent derive, want (1,true)", got, ok)
					return
				}
				if !base.Has("b") {
					t.Errorf("Has(b) = false during concurrent derive, want true")
					return
				}
			}
		})
	}

	go func() {
		v := base
		for i := range numDerivations {
			v = v.Set("dynamic", i)
		}
	}()

	for i := range numDerivations {
		_ = base.Set("other", i)
	}
	close(stop)
	wg.Wait()

	if got, ok := base.Get("a"); got != 1 || !ok {
		t.Fatalf("base changed after concurrent derivation: Get(a) = (%d,%v), want (1,true)", got, ok)
	}
}

// TestSetConcurrentReadDuringDerive is the Set analogue of
// TestListConcurrentReadDuringDerive, deriving with Add instead of Append.
func TestSetConcurrentReadDuringDerive(t *testing.T) {
	t.Parallel()

	base := immutable.SetOf(1, 2, 3)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for range numReaders {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if !base.Has(2) {
					t.Errorf("Has(2) = false during concurrent derive, want true")
					return
				}
				if base.Len() != 3 {
					t.Errorf("Len() = %d during concurrent derive, want 3", base.Len())
					return
				}
			}
		})
	}

	go func() {
		v := base
		for i := range numDerivations {
			v = v.Add(100 + i)
		}
	}()

	for i := range numDerivations {
		_ = base.Add(1000 + i)
	}
	close(stop)
	wg.Wait()

	if base.Len() != 3 {
		t.Fatalf("base changed after concurrent derivation: Len() = %d, want 3", base.Len())
	}
}

// TestSetAlgebraConcurrentReadDuringDerive proves that reading shared Set
// operands is unaffected by writer goroutines computing Union, Intersection,
// and Difference from them, for operands with independent seeds and for
// operands derived from one lineage. Under -race none of this may report a
// data race, and every reader and writer must keep seeing the operands'
// original contents.
func TestSetAlgebraConcurrentReadDuringDerive(t *testing.T) {
	t.Parallel()

	shared := immutable.SetOf(1, 2, 3, 4)
	lineages := []struct {
		name string
		a, b immutable.Set[int]
	}{
		{"independent", shared, immutable.SetOf(3, 4, 5, 6)},
		{"derived", shared, shared.Remove(1, 2).Add(5, 6)},
	}
	for _, lin := range lineages {
		t.Run(lin.name, func(t *testing.T) {
			t.Parallel()
			a, b := lin.a, lin.b
			wantA, wantB := sortedItems(a), sortedItems(b)

			var wg sync.WaitGroup
			stop := make(chan struct{})
			for range numReaders {
				wg.Go(func() {
					for {
						select {
						case <-stop:
							return
						default:
						}
						if a.Len() != 4 || !a.Has(1) || b.Len() != 4 || !b.Has(6) {
							t.Errorf("operands changed during concurrent derive: a=%v b=%v", a.Items(), b.Items())
							return
						}
					}
				})
			}

			var writers sync.WaitGroup
			for range 2 {
				writers.Go(func() {
					for range numDerivations {
						checkSetAlgebraLens(t, a, b)
					}
				})
			}
			writers.Wait()
			close(stop)
			wg.Wait()

			if got := sortedItems(a); !slices.Equal(got, wantA) {
				t.Fatalf("operand a changed after concurrent derivation: got %v, want %v", got, wantA)
			}
			if got := sortedItems(b); !slices.Equal(got, wantB) {
				t.Fatalf("operand b changed after concurrent derivation: got %v, want %v", got, wantB)
			}
		})
	}
}

// checkSetAlgebraLens computes every set-algebra operation over a and b, whose
// four elements overlap in exactly two, and checks each result's size.
func checkSetAlgebraLens(t *testing.T, a, b immutable.Set[int]) {
	t.Helper()

	lens := []struct {
		op        string
		got, want int
	}{
		{"a.Union(b)", a.Union(b).Len(), 6},
		{"b.Union(a)", b.Union(a).Len(), 6},
		{"a.Intersection(b)", a.Intersection(b).Len(), 2},
		{"b.Intersection(a)", b.Intersection(a).Len(), 2},
		{"a.Difference(b)", a.Difference(b).Len(), 2},
		{"b.Difference(a)", b.Difference(a).Len(), 2},
	}
	for _, l := range lens {
		if l.got != l.want {
			t.Errorf("%s.Len() = %d during concurrent derive, want %d", l.op, l.got, l.want)
		}
	}
}
