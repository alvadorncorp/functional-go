package action_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func TestTapPreservesValueAfterSuccessfulEffect(t *testing.T) {
	events := make([]string, 0, 2)
	var seen context.Context
	source := action.Do(func(ctx context.Context) (int, error) {
		events = append(events, "source")
		return 42, nil
	})
	program := action.Tap(source, func(value int) action.Action[struct{}] {
		if value != 42 {
			t.Fatalf("effect value = %d, want 42", value)
		}
		return action.Do(func(ctx context.Context) (struct{}, error) {
			events = append(events, "effect")
			seen = ctx
			return struct{}{}, nil
		})
	})
	if len(events) != 0 {
		t.Fatalf("construction events = %v, want none", events)
	}

	value, err := action.Run(t.Context(), program)
	if err != nil || value != 42 || seen != t.Context() {
		t.Fatalf("Run() = (%d, %v), context preserved=%v", value, err, seen == t.Context())
	}
	if len(events) != 2 || events[0] != "source" || events[1] != "effect" {
		t.Fatalf("events = %v, want [source effect]", events)
	}
}

func TestTapShortCircuitsAndPreservesExactErrors(t *testing.T) {
	sourceErr := errors.New("source")
	effectErr := errors.New("effect")
	effectCalls := 0

	_, err := action.Run(t.Context(), action.Tap(action.Fail[int](sourceErr), func(int) action.Action[struct{}] {
		effectCalls++
		return action.Pure(struct{}{})
	}))
	if err != sourceErr || effectCalls != 0 {
		t.Fatalf("source failure = (%v, calls=%d), want identical error and zero calls", err, effectCalls)
	}

	_, err = action.Run(t.Context(), action.Tap(action.Pure(42), func(int) action.Action[struct{}] {
		effectCalls++
		return action.Fail[struct{}](effectErr)
	}))
	if err != effectErr || effectCalls != 1 {
		t.Fatalf("effect failure = (%v, calls=%d), want identical error and one call", err, effectCalls)
	}
}

func TestTapIsReExecutable(t *testing.T) {
	sourceRuns := 0
	effectRuns := 0
	program := action.Tap(
		action.Do(func(context.Context) (int, error) { sourceRuns++; return sourceRuns, nil }),
		func(int) action.Action[struct{}] {
			return action.Do(func(context.Context) (struct{}, error) {
				effectRuns++
				return struct{}{}, nil
			})
		},
	)
	for range 2 {
		if _, err := action.Run(t.Context(), program); err != nil {
			t.Fatal(err)
		}
	}
	if sourceRuns != 2 || effectRuns != 2 {
		t.Fatalf("runs = (%d, %d), want (2, 2)", sourceRuns, effectRuns)
	}
}

func TestTapRejectsNilArgumentsAndNilEffectAction(t *testing.T) {
	var nilAction action.Action[int]
	var nilEffect func(int) action.Action[struct{}]
	assertPanics(t, "workflow/action: nil action", func() { action.Tap(nilAction, func(int) action.Action[struct{}] { return nil }) })
	assertPanics(t, "workflow/action: nil effect", func() { action.Tap(action.Pure(1), nilEffect) })

	program := action.Tap(action.Pure(1), func(int) action.Action[struct{}] { return nil })
	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), program)
	})
}
