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
	assertPanics(t, "workflow/action: nil action", func() {
		action.Tap(nilAction, func(int) action.Action[struct{}] { return nil })
	})
	assertPanics(t, "workflow/action: nil effect", func() { action.Tap(action.Pure(1), nilEffect) })

	program := action.Tap(action.Pure(1), func(int) action.Action[struct{}] { return nil })
	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), program)
	})
}

func TestEnsureResultMatrix(t *testing.T) {
	primaryErr := errors.New("primary")
	finalizerErr := errors.New("finalizer")
	tests := []struct {
		name         string
		primaryErr   error
		finalizerErr error
		wantValue    int
		wantPrimary  bool
		wantFinal    bool
	}{
		{name: "both succeed", wantValue: 42},
		{name: "source fails", primaryErr: primaryErr, wantPrimary: true},
		{name: "finalizer fails", finalizerErr: finalizerErr, wantFinal: true},
		{name: "both fail", primaryErr: primaryErr, finalizerErr: finalizerErr, wantPrimary: true, wantFinal: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalizerRuns := 0
			source := action.Do(func(context.Context) (int, error) { return 42, tt.primaryErr })
			finalizer := action.Do(func(context.Context) (struct{}, error) {
				finalizerRuns++
				return struct{}{}, tt.finalizerErr
			})
			value, err := action.Run(t.Context(), action.Ensure(source, finalizer))
			if value != tt.wantValue || finalizerRuns != 1 {
				t.Fatalf("value=%d finalizerRuns=%d, want %d and 1", value, finalizerRuns, tt.wantValue)
			}
			if errors.Is(err, primaryErr) != tt.wantPrimary || errors.Is(err, finalizerErr) != tt.wantFinal {
				t.Fatalf("error=%v, want primary=%v finalizer=%v", err, tt.wantPrimary, tt.wantFinal)
			}
			if tt.primaryErr != nil && tt.finalizerErr == nil && err != primaryErr {
				t.Fatalf("source-only error = %v, want identical %v", err, primaryErr)
			}
			if tt.primaryErr == nil && tt.finalizerErr != nil && err != finalizerErr {
				t.Fatalf("finalizer-only error = %v, want identical %v", err, finalizerErr)
			}
		})
	}
}

func TestEnsureIsLazyOrderedReExecutableAndContextPreserving(t *testing.T) {
	events := make([]string, 0, 4)
	var seen []context.Context
	source := action.Do(func(ctx context.Context) (int, error) {
		events = append(events, "source")
		seen = append(seen, ctx)
		return 42, nil
	})
	finalizer := action.Do(func(ctx context.Context) (struct{}, error) {
		events = append(events, "finalizer")
		seen = append(seen, ctx)
		return struct{}{}, nil
	})
	program := action.Ensure(source, finalizer)
	if len(events) != 0 {
		t.Fatalf("construction events = %v, want none", events)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for range 2 {
		if _, err := action.Run(ctx, program); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"source", "finalizer", "source", "finalizer"}
	for i := range want {
		if events[i] != want[i] || seen[i] != ctx {
			t.Fatalf("events=%v context[%d] preserved=%v", events, i, seen[i] == ctx)
		}
	}
}

func TestEnsureDoesNotFinalizeAfterPanic(t *testing.T) {
	finalizerRuns := 0
	program := action.Ensure(
		action.Do(func(context.Context) (int, error) { panic("source panic") }),
		action.Do(func(context.Context) (struct{}, error) { finalizerRuns++; return struct{}{}, nil }),
	)
	assertPanics(t, "source panic", func() { _, _ = action.Run(t.Context(), program) })
	if finalizerRuns != 0 {
		t.Fatalf("finalizer runs = %d, want 0 after panic", finalizerRuns)
	}
}

func TestEnsureRejectsNilActions(t *testing.T) {
	var nilInt action.Action[int]
	var nilUnit action.Action[struct{}]
	assertPanics(t, "workflow/action: nil action", func() { action.Ensure(nilInt, action.Pure(struct{}{})) })
	assertPanics(t, "workflow/action: nil action", func() { action.Ensure(action.Pure(1), nilUnit) })
}
