package lazy_test

import (
	"sync/atomic"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/lazy"
)

func TestMapAndFlatMapStayLazyAndMemoized(t *testing.T) {
	t.Parallel()
	var sourceCalls, mapCalls, flatMapCalls atomic.Int32
	source := lazy.New(func() int { sourceCalls.Add(1); return 5 })
	mapped := lazy.Map(source, func(v int) int { mapCalls.Add(1); return v * 2 })
	flatMapped := lazy.FlatMap(mapped, func(v int) *lazy.Lazy[int] {
		flatMapCalls.Add(1)
		return lazy.Of(v * 3)
	})
	if sourceCalls.Load()+mapCalls.Load()+flatMapCalls.Load() != 0 {
		t.Fatal("composition evaluated eagerly")
	}
	if got := flatMapped.Get(); got != 30 {
		t.Fatalf("Get = %d", got)
	}
	if got := flatMapped.Get(); got != 30 {
		t.Fatalf("second Get = %d", got)
	}
	if sourceCalls.Load() != 1 || mapCalls.Load() != 1 || flatMapCalls.Load() != 1 {
		t.Fatalf("calls = source:%d map:%d flat:%d", sourceCalls.Load(), mapCalls.Load(), flatMapCalls.Load())
	}
}

func TestMapComposition(t *testing.T) {
	t.Parallel()
	value := lazy.Map(lazy.Map(lazy.Of(3), func(v int) int { return v + 1 }), func(v int) int { return v * 2 })
	if got := value.Get(); got != 8 {
		t.Fatalf("Get = %d", got)
	}
}
