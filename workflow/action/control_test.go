package action_test

import (
	"context"
	"errors"
	"testing"

	"go.alvadorn.org/functional/workflow/action"
)

const fallbackValue = 7

func TestRecoverIsNotCalledOnSuccess(t *testing.T) {
	t.Parallel()

	recoveries := 0
	recovered := action.Recover(action.Pure(answer), func(error) action.Action[int] {
		recoveries++
		return action.Pure(fallbackValue)
	})

	value, err := action.Run(t.Context(), recovered)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != answer {
		t.Fatalf("Run() value = %d, want %d", value, answer)
	}
	if recoveries != 0 {
		t.Fatalf("recovery calls = %d, want 0", recoveries)
	}
}

func TestRecoverReplacesTheFailure(t *testing.T) {
	t.Parallel()

	var seen error
	recovered := action.Recover(action.Fail[int](errBoom), func(err error) action.Action[int] {
		seen = err
		return action.Pure(fallbackValue)
	})

	value, err := action.Run(t.Context(), recovered)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != fallbackValue {
		t.Fatalf("Run() value = %d, want %d", value, fallbackValue)
	}
	if !errors.Is(seen, errBoom) {
		t.Fatalf("recovery received %v, want errBoom", seen)
	}
}

func TestRecoverMayRethrow(t *testing.T) {
	t.Parallel()

	recovered := action.Recover(action.Fail[int](errBoom), func(err error) action.Action[int] {
		return action.Fail[int](err)
	})

	if _, err := action.Run(t.Context(), recovered); !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want errBoom", err)
	}
}

func TestRecoverSeesCancellationAsAnOrdinaryError(t *testing.T) {
	t.Parallel()

	var seen error
	failing := action.Do(func(ctx context.Context) (int, error) {
		return 0, ctx.Err()
	})
	recovered := action.Recover(failing, func(err error) action.Action[int] {
		seen = err
		return action.Fail[int](err)
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := action.Run(ctx, recovered); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if !errors.Is(seen, context.Canceled) {
		t.Fatalf("recovery received %v, want context.Canceled unchanged", seen)
	}
}

func TestRecoverDoesNotCatchPanics(t *testing.T) {
	t.Parallel()

	panicking := action.Do(func(context.Context) (int, error) {
		panic("workflow/action_test: deliberate panic")
	})
	recovered := action.Recover(panicking, func(error) action.Action[int] {
		return action.Pure(fallbackValue)
	})

	assertPanics(t, "workflow/action_test: deliberate panic", func() {
		_, _ = action.Run(t.Context(), recovered)
	})
}

func TestRecoverPanicsOnNilArguments(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil action", func() {
		_ = action.Recover(action.Action[int](nil), func(error) action.Action[int] {
			return action.Pure(fallbackValue)
		})
	})
	assertPanics(t, "workflow/action: nil recovery", func() {
		_ = action.Recover[int](action.Pure(answer), nil)
	})
}

func TestRecoverPanicsWhenRecoveryReturnsNilAction(t *testing.T) {
	t.Parallel()

	recovered := action.Recover(action.Fail[int](errBoom), func(error) action.Action[int] {
		return nil
	})

	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), recovered)
	})
}
