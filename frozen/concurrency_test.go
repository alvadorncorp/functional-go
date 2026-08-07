package frozen_test

import (
	"sync"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

// numReaders is how many goroutines TestSliceConcurrentReads and
// TestMapConcurrentReads run in parallel. Run this file with -race: Slice and
// Map never write after construction, so concurrent reads must never race.
const numReaders = 50

func TestSliceConcurrentReads(t *testing.T) {
	t.Parallel()

	s := frozen.NewSlice(1, 2, 3, 4, 5)

	var wg sync.WaitGroup
	for range numReaders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range s.Values() {
			}
			for range s.All() {
			}
			_ = s.Items()
			_ = s.At(0)
			_ = s.Len()
		}()
	}
	wg.Wait()
}

func TestMapConcurrentReads(t *testing.T) {
	t.Parallel()

	m := frozen.NewMap(map[string]int{"a": 1, "b": 2, "c": 3})

	var wg sync.WaitGroup
	for range numReaders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range m.All() {
			}
			_ = m.Items()
			_, _ = m.Get("a")
			_ = m.Has("b")
			_ = m.Len()
		}()
	}
	wg.Wait()
}
