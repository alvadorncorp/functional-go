package immutable_test

import (
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
