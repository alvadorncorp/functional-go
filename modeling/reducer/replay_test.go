package reducer_test

import (
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
	"github.com/alvadorncorp/gopackages/modeling/reducer"
)

type replayFunc[S, E any] func(
	S,
	immutable.List[E],
	reducer.Reducer[S, E],
) S

var (
	_ replayFunc[int, int]              = reducer.Replay[int, int]
	_ replayFunc[[]int, map[string]int] = reducer.Replay[[]int, map[string]int]
)

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

	got := reducer.Replay(0, immutable.ListOf(1, 2, 3), reduce)

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

	got := reducer.Replay(42, immutable.List[int]{}, reduce)

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

	got := reducer.Replay(initial, immutable.List[map[string]int]{}, reduce)

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
		immutable.ListOf(map[string]int{"value": 1}),
		reduce,
	)
}

func TestReplaySupportsNonComparableStateAndEvent(t *testing.T) {
	t.Parallel()

	got := reducer.Replay(
		[]int{1},
		immutable.ListOf(map[string]int{"value": 2}),
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

	left := immutable.ListOf(1, 2)
	right := immutable.ListOf(3, 4)
	joined := left.Append(right.Items()...)

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

func TestReplayIsDeterministicForDeterministicReducerAndPreservesEvents(t *testing.T) {
	t.Parallel()

	events := immutable.ListOf(2, 3, 5)
	before := events.Items()
	reduce := func(state, event int) int {
		return state*31 + event
	}

	first := reducer.Replay(7, events, reduce)
	second := reducer.Replay(7, events, reduce)

	if first != second {
		t.Fatalf("deterministic Replay results = %d and %d", first, second)
	}
	if !slices.Equal(events.Items(), before) {
		t.Fatalf("Replay changed events: got %v, want %v", events.Items(), before)
	}
}
