package reducer_test

import (
	"iter"
	"slices"
	"strconv"
	"testing"

	"go.alvadorn.org/functional/modeling/reducer"
)

type replayFunc[S, E any] func(
	S,
	iter.Seq[E],
	reducer.Reducer[S, E],
) S

var (
	_ replayFunc[int, int]              = reducer.Replay[int, int]
	_ replayFunc[[]int, map[string]int] = reducer.Replay[[]int, map[string]int]
)

func seqOf[E any](events ...E) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, event := range events {
			if !yield(event) {
				return
			}
		}
	}
}

func emptySeq[E any]() iter.Seq[E] {
	return func(func(E) bool) {}
}

func concatSeq[E any](left, right iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for event := range left {
			if !yield(event) {
				return
			}
		}
		for event := range right {
			if !yield(event) {
				return
			}
		}
	}
}

func decimalReducer(state, event int) int {
	return state*10 + event
}

func TestReplayThreadsStateInEncounterOrder(t *testing.T) {
	t.Parallel()

	var calls []int
	reduce := func(state, event int) int {
		calls = append(calls, event)
		return decimalReducer(state, event)
	}

	got := reducer.Replay(0, seqOf(1, 2, 3), reduce)

	if got != 123 {
		t.Fatalf("Replay result = %d, want 123", got)
	}
	if !slices.Equal(calls, []int{1, 2, 3}) {
		t.Fatalf("Replay calls = %v, want [1 2 3]", calls)
	}
}

func TestReplayEmptyReturnsInitialWithoutCallingReducer(t *testing.T) {
	t.Parallel()

	calls := 0
	reduce := func(state, event int) int {
		calls++
		return state + event
	}

	got := reducer.Replay(42, emptySeq[int](), reduce)

	if got != 42 {
		t.Fatalf("Replay empty result = %d, want 42", got)
	}
	if calls != 0 {
		t.Fatalf("Replay empty calls = %d, want 0", calls)
	}
}

func TestReplayEmptyAcceptsNilReducer(t *testing.T) {
	t.Parallel()

	initial := []int{1, 2}
	var reduce reducer.Reducer[[]int, map[string]int]

	got := reducer.Replay(initial, emptySeq[map[string]int](), reduce)

	if !slices.Equal(got, initial) {
		t.Fatalf("Replay empty with nil reducer = %v, want %v", got, initial)
	}
}

func TestReplayNonEmptyNilReducerHasOrdinaryFunctionPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("Replay with non-empty events and nil reducer did not panic")
		}
	}()

	var reduce reducer.Reducer[[]int, map[string]int]
	_ = reducer.Replay(
		[]int{},
		seqOf(map[string]int{"value": 1}),
		reduce,
	)
}

func TestReplaySupportsNonComparableStateAndEvent(t *testing.T) {
	t.Parallel()

	got := reducer.Replay(
		[]int{1},
		seqOf(map[string]int{"value": 2}),
		func(state []int, event map[string]int) []int {
			return append(state, event["value"])
		},
	)

	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("Replay non-comparable result = %v, want [1 2]", got)
	}
}

func TestReplayConcatenationLaw(t *testing.T) {
	t.Parallel()

	left := seqOf(1, 2)
	right := seqOf(3, 4)
	joined := concatSeq(left, right)

	direct := reducer.Replay(0, joined, decimalReducer)
	staged := reducer.Replay(
		reducer.Replay(0, left, decimalReducer),
		right,
		decimalReducer,
	)

	if direct != staged {
		t.Fatalf("concatenated Replay = %d, staged Replay = %d", direct, staged)
	}
}

func TestReplayIsDeterministicForDeterministicReducer(t *testing.T) {
	t.Parallel()

	events := []int{2, 3, 5}
	reduce := func(state, event int) int {
		return state*31 + event
	}

	first := reducer.Replay(7, seqOf(events...), reduce)
	second := reducer.Replay(7, seqOf(events...), reduce)

	if first != second {
		t.Fatalf("deterministic Replay results = %d and %d", first, second)
	}
	if !slices.Equal(events, []int{2, 3, 5}) {
		t.Fatalf("Replay changed source slice: got %v", events)
	}
}

func TestReplayIsSinglePass(t *testing.T) {
	t.Parallel()

	passes := 0
	events := func(yield func(int) bool) {
		passes++
		if passes > 1 {
			t.Fatal("Replay ranged over events more than once")
		}
		for _, event := range []int{1, 2} {
			if !yield(event) {
				return
			}
		}
	}

	got := reducer.Replay(0, events, decimalReducer)
	if got != 12 {
		t.Fatalf("Replay single-pass result = %d, want 12", got)
	}
	if passes != 1 {
		t.Fatalf("Replay passes = %d, want 1", passes)
	}
}

var (
	replaySumSink        int
	replayBenchmarkSizes = [...]int{0, 1, 10, 100, 1000, 10000}
)

func benchmarkIntEvents(size int) iter.Seq[int] {
	events := make([]int, size)
	for index := range events {
		events[index] = index
	}
	return seqOf(events...)
}

func BenchmarkReplayInt(b *testing.B) {
	for _, size := range replayBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := benchmarkIntEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				replaySumSink = reducer.Replay(0, events, func(total, event int) int {
					return total + event
				})
			}
		})
	}
}
