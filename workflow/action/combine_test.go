package action_test

import (
	"context"
	"errors"
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
	//nolint:staticcheck // private typed key verifies exact context propagation.
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
}
