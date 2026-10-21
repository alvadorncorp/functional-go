package action_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

type combineContextKey struct{}

func TestZipRunsSequentiallyAndPreservesValuesAndContext(t *testing.T) {
	events := make([]string, 0, 2)
	var seen []context.Context
	first := action.Do(func(ctx context.Context) (int, error) {
		events = append(events, "first")
		seen = append(seen, ctx)
		return 40, nil
	})
	second := action.Do(func(ctx context.Context) (string, error) {
		events = append(events, "second")
		seen = append(seen, ctx)
		marker, _ := ctx.Value(combineContextKey{}).(string)
		return marker, nil
	})

	program := action.Zip(first, second)
	if len(events) != 0 {
		t.Fatalf("construction events = %v, want none", events)
	}
	ctx := context.WithValue(t.Context(), combineContextKey{}, "two")
	pair, err := action.Run(ctx, program)
	if err != nil || pair.First != 40 || pair.Second != "two" {
		t.Fatalf("Run() = (%+v, %v), want ({40 two}, nil)", pair, err)
	}
	if len(events) != 2 || events[0] != "first" || events[1] != "second" {
		t.Fatalf("events = %v, want [first second]", events)
	}
	if len(seen) != 2 || seen[0] != ctx || seen[1] != ctx {
		t.Fatalf("contexts = %v, want identical Run context twice", seen)
	}
}

func TestZipShortCircuitsAndPreservesErrors(t *testing.T) {
	firstErr := errors.New("first")
	secondErr := errors.New("second")
	secondRuns := 0
	_, err := action.Run(t.Context(), action.Zip(
		action.Fail[int](firstErr),
		action.Do(func(context.Context) (string, error) { secondRuns++; return "", nil }),
	))
	if err != firstErr || secondRuns != 0 {
		t.Fatalf("first failure = (%v, runs=%d), want identical error and zero runs", err, secondRuns)
	}

	_, err = action.Run(t.Context(), action.Zip(action.Pure(1), action.Fail[string](secondErr)))
	if err != secondErr {
		t.Fatalf("second failure = %v, want identical %v", err, secondErr)
	}
}

func TestBothCombinesOnlyAfterTwoSuccesses(t *testing.T) {
	events := make([]string, 0, 3)
	first := action.Do(func(context.Context) (int, error) { events = append(events, "first"); return 20, nil })
	second := action.Do(func(context.Context) (int, error) { events = append(events, "second"); return 22, nil })
	program := action.Both(first, second, func(a, b int) int {
		events = append(events, "combine")
		return a + b
	})

	value, err := action.Run(t.Context(), program)
	if err != nil || value != 42 {
		t.Fatalf("Run() = (%d, %v), want (42, nil)", value, err)
	}
	want := []string{"first", "second", "combine"}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestBothShortCircuitsBeforeCombiner(t *testing.T) {
	wantErr := errors.New("failed")
	secondRuns := 0
	combineRuns := 0
	combine := func(a, b int) int { combineRuns++; return a + b }

	_, err := action.Run(t.Context(), action.Both(
		action.Fail[int](wantErr),
		action.Do(func(context.Context) (int, error) { secondRuns++; return 0, nil }),
		combine,
	))
	if err != wantErr || secondRuns != 0 || combineRuns != 0 {
		t.Fatalf("first failure = (%v, second=%d, combine=%d)", err, secondRuns, combineRuns)
	}

	_, err = action.Run(t.Context(), action.Both(action.Pure(1), action.Fail[int](wantErr), combine))
	if err != wantErr || combineRuns != 0 {
		t.Fatalf("second failure = (%v, combine=%d)", err, combineRuns)
	}
}

func TestZipAndBothAreLazyAndReExecutable(t *testing.T) {
	runs := 0
	source := func(value int) action.Action[int] {
		return action.Do(func(context.Context) (int, error) { runs++; return value, nil })
	}
	zipped := action.Zip(source(1), source(2))
	both := action.Both(source(3), source(4), func(a, b int) int { return a + b })
	if runs != 0 {
		t.Fatalf("construction runs = %d, want 0", runs)
	}
	for range 2 {
		if _, err := action.Run(t.Context(), zipped); err != nil {
			t.Fatal(err)
		}
		if _, err := action.Run(t.Context(), both); err != nil {
			t.Fatal(err)
		}
	}
	if runs != 8 {
		t.Fatalf("runs = %d, want 8", runs)
	}
}

func TestCombinationRejectsNilArguments(t *testing.T) {
	var nilAction action.Action[int]
	var nilCombiner func(int, int) int

	assertPanics(t, "workflow/action: nil action", func() { action.Zip(nilAction, action.Pure(1)) })
	assertPanics(t, "workflow/action: nil action", func() { action.Zip(action.Pure(1), nilAction) })
	assertPanics(t, "workflow/action: nil action", func() {
		action.Both(nilAction, action.Pure(1), func(a, b int) int { return a + b })
	})
	assertPanics(t, "workflow/action: nil action", func() {
		action.Both(action.Pure(1), nilAction, func(a, b int) int { return a + b })
	})
	assertPanics(t, "workflow/action: nil combiner", func() {
		action.Both(action.Pure(1), action.Pure(2), nilCombiner)
	})
	assertPanics(t, "workflow/action: nil binder", func() { action.Traverse[int, int]([]int{1}, nil) })
	assertPanics(t, "workflow/action: nil binder", func() { action.Traverse[int, int](nil, nil) })
}

func TestTraverseRunsInIndexOrderInterleavedWithTheRunContext(t *testing.T) {
	var events []string
	var seen []context.Context
	program := action.Traverse([]string{"a", "b", "c"}, func(item string) action.Action[string] {
		events = append(events, "bind "+item)
		return action.Do(func(ctx context.Context) (string, error) {
			events = append(events, "run "+item)
			seen = append(seen, ctx)
			return item + item, nil
		})
	})
	if len(events) != 0 {
		t.Fatalf("construction events = %v, want none", events)
	}

	ctx := context.WithValue(t.Context(), combineContextKey{}, "traverse")
	values, err := action.Run(ctx, program)
	if err != nil || !slices.Equal(values, []string{"aa", "bb", "cc"}) {
		t.Fatalf("Run() = (%v, %v), want ([aa bb cc], nil)", values, err)
	}
	want := []string{"bind a", "run a", "bind b", "run b", "bind c", "run c"}
	if !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(seen) != 3 || seen[0] != ctx || seen[1] != ctx || seen[2] != ctx {
		t.Fatalf("contexts = %v, want identical Run context three times", seen)
	}
}

func TestTraverseStopsAtTheFirstFailure(t *testing.T) {
	wantErr := errors.New("second failed")
	var bound []int
	program := action.Traverse([]int{1, 2, 3}, func(item int) action.Action[int] {
		bound = append(bound, item)
		if item == 2 {
			return action.Fail[int](wantErr)
		}
		return action.Pure(item)
	})

	values, err := action.Run(t.Context(), program)
	if err != wantErr || values != nil {
		t.Fatalf("Run() = (%v, %v), want (nil, identical %v)", values, err, wantErr)
	}
	if !slices.Equal(bound, []int{1, 2}) {
		t.Fatalf("bound items = %v, want [1 2]", bound)
	}
}

func TestTraverseOfNoItemsReturnsAnEmptyNonNilSlice(t *testing.T) {
	binds := 0
	binder := func(item int) action.Action[int] { binds++; return action.Pure(item) }

	for _, items := range [][]int{nil, {}} {
		values, err := action.Run(t.Context(), action.Traverse(items, binder))
		if err != nil || values == nil || len(values) != 0 {
			t.Fatalf("Run(%#v) = (%#v, %v), want (non-nil empty slice, nil)", items, values, err)
		}
	}
	if binds != 0 {
		t.Fatalf("binder calls = %d, want 0", binds)
	}
}

func TestTraverseIsLazyAndReExecutableWithoutAliasing(t *testing.T) {
	runs := 0
	program := action.Traverse([]int{1, 2}, func(item int) action.Action[int] {
		return action.Do(func(context.Context) (int, error) { runs++; return item * 10, nil })
	})
	if runs != 0 {
		t.Fatalf("construction runs = %d, want 0", runs)
	}

	first, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatal(err)
	}
	first[0] = -1
	second, err := action.Run(t.Context(), program)
	if err != nil || !slices.Equal(second, []int{10, 20}) {
		t.Fatalf("second Run() = (%v, %v), want ([10 20], nil)", second, err)
	}
	if !slices.Equal(first, []int{-1, 20}) {
		t.Fatalf("first result after second Run = %v, want [-1 20] (no shared backing array)", first)
	}
	if runs != 4 {
		t.Fatalf("runs = %d, want 4", runs)
	}
}

func TestTraverseCopiesItemsAtConstruction(t *testing.T) {
	items := []int{1, 2}
	program := action.Traverse(items, action.Pure[int])
	items[0] = 99

	values, err := action.Run(t.Context(), program)
	if err != nil || !slices.Equal(values, []int{1, 2}) {
		t.Fatalf("Run() = (%v, %v), want ([1 2], nil)", values, err)
	}
}

func TestTraverseForwardsACancelledContextUnchanged(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var seen []context.Context
	program := action.Traverse([]int{1, 2}, func(item int) action.Action[int] {
		return action.Do(func(ctx context.Context) (int, error) {
			seen = append(seen, ctx)
			return item, nil
		})
	})

	values, err := action.Run(ctx, program)
	if err != nil || !slices.Equal(values, []int{1, 2}) {
		t.Fatalf("Run() = (%v, %v), want ([1 2], nil): Traverse must not check the context", values, err)
	}
	if len(seen) != 2 || seen[0] != ctx || seen[1] != ctx {
		t.Fatalf("contexts = %v, want the cancelled Run context twice", seen)
	}
}

func TestTraverseRejectsANilActionAtRun(t *testing.T) {
	program := action.Traverse([]int{1}, func(int) action.Action[int] { return nil })
	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), program)
	})
}
