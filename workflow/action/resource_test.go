package action_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func TestBracketResultMatrix(t *testing.T) {
	acquireErr := errors.New("acquire")
	useErr := errors.New("use")
	releaseErr := errors.New("release")
	tests := []struct {
		name       string
		acquireErr error
		useErr     error
		releaseErr error
		wantValue  int
		wantUse    int
		wantRel    int
	}{
		{name: "acquire fails", acquireErr: acquireErr},
		{name: "all succeed", wantValue: 42, wantUse: 1, wantRel: 1},
		{name: "use fails", useErr: useErr, wantUse: 1, wantRel: 1},
		{name: "release fails", releaseErr: releaseErr, wantUse: 1, wantRel: 1},
		{name: "use and release fail", useErr: useErr, releaseErr: releaseErr, wantUse: 1, wantRel: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := make([]string, 0, 3)
			useRuns := 0
			releaseRuns := 0
			acquire := action.Do(func(context.Context) (string, error) {
				events = append(events, "acquire")
				return "resource", tt.acquireErr
			})
			program := action.Bracket(acquire, func(resource string) action.Action[int] {
				if resource != "resource" {
					t.Fatalf("use resource = %q", resource)
				}
				return action.Do(func(context.Context) (int, error) {
					useRuns++
					events = append(events, "use")
					return 42, tt.useErr
				})
			}, func(resource string) action.Action[struct{}] {
				return action.Do(func(context.Context) (struct{}, error) {
					releaseRuns++
					events = append(events, "release")
					return struct{}{}, tt.releaseErr
				})
			})
			value, err := action.Run(t.Context(), program)
			if value != tt.wantValue || useRuns != tt.wantUse || releaseRuns != tt.wantRel {
				t.Fatalf("value=%d use=%d release=%d events=%v", value, useRuns, releaseRuns, events)
			}
			for _, candidate := range []error{tt.acquireErr, tt.useErr, tt.releaseErr} {
				if candidate != nil && !errors.Is(err, candidate) {
					t.Fatalf("error %v does not preserve %v", err, candidate)
				}
			}
			if tt.acquireErr != nil && err != acquireErr {
				t.Fatalf("acquire error = %v, want identical %v", err, acquireErr)
			}
			if tt.useErr != nil && tt.releaseErr == nil && err != useErr {
				t.Fatalf("use error = %v, want identical %v", err, useErr)
			}
			if tt.useErr == nil && tt.releaseErr != nil && err != releaseErr {
				t.Fatalf("release error = %v, want identical %v", err, releaseErr)
			}
		})
	}
}

func TestBracketPropagatesCanceledContextAndReexecutes(t *testing.T) {
	var seen []context.Context
	events := make([]string, 0, 6)
	program := action.Bracket(
		action.Do(func(ctx context.Context) (string, error) {
			seen = append(seen, ctx)
			events = append(events, "acquire")
			return "r", nil
		}),
		func(string) action.Action[int] {
			return action.Do(func(ctx context.Context) (int, error) {
				seen = append(seen, ctx)
				events = append(events, "use")
				return 42, nil
			})
		},
		func(string) action.Action[struct{}] {
			return action.Do(func(ctx context.Context) (struct{}, error) {
				seen = append(seen, ctx)
				events = append(events, "release")
				return struct{}{}, nil
			})
		},
	)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for range 2 {
		if _, err := action.Run(ctx, program); err != nil {
			t.Fatal(err)
		}
	}
	for i, got := range seen {
		if got != ctx || got.Err() != context.Canceled {
			t.Fatalf("context[%d] not identical canceled Run context", i)
		}
	}
	want := []string{"acquire", "use", "release", "acquire", "use", "release"}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestBracketDoesNotReleaseAfterPanic(t *testing.T) {
	releaseRuns := 0
	program := action.Bracket(
		action.Pure("resource"),
		func(string) action.Action[int] {
			return action.Do(func(context.Context) (int, error) { panic("use panic") })
		},
		func(string) action.Action[struct{}] { releaseRuns++; return action.Pure(struct{}{}) },
	)
	assertPanics(t, "use panic", func() { _, _ = action.Run(t.Context(), program) })
	if releaseRuns != 0 {
		t.Fatalf("release callback runs = %d, want 0", releaseRuns)
	}
}

func TestBracketRejectsNilArgumentsAndCallbackResults(t *testing.T) {
	var nilAcquire action.Action[string]
	var nilUse func(string) action.Action[int]
	var nilRelease func(string) action.Action[struct{}]
	validUse := func(string) action.Action[int] { return action.Pure(1) }
	validRelease := func(string) action.Action[struct{}] { return action.Pure(struct{}{}) }

	assertPanics(t, "workflow/action: nil action", func() { action.Bracket(nilAcquire, validUse, validRelease) })
	assertPanics(t, "workflow/action: nil use", func() { action.Bracket(action.Pure("r"), nilUse, validRelease) })
	assertPanics(t, "workflow/action: nil release", func() { action.Bracket(action.Pure("r"), validUse, nilRelease) })
	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), action.Bracket(
			action.Pure("r"), func(string) action.Action[int] { return nil }, validRelease,
		))
	})
	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), action.Bracket(
			action.Pure("r"), validUse, func(string) action.Action[struct{}] { return nil },
		))
	})
}

func TestAcquireUseReleaseMatchesBracket(t *testing.T) {
	acquire := action.Pure("resource")
	use := func(string) action.Action[int] { return action.Pure(42) }
	release := func(string) action.Action[struct{}] { return action.Pure(struct{}{}) }
	bracketValue, bracketErr := action.Run(t.Context(), action.Bracket(acquire, use, release))
	aliasValue, aliasErr := action.Run(t.Context(), action.AcquireUseRelease(acquire, use, release))
	if bracketValue != aliasValue || bracketErr != aliasErr {
		t.Fatalf("Bracket=(%d,%v) alias=(%d,%v)", bracketValue, bracketErr, aliasValue, aliasErr)
	}
}
