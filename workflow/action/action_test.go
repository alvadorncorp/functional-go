package action_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

const answer = 42

var errBoom = errors.New("boom")

type contextKey struct{}

func TestActionConstructionExecutesNothing(t *testing.T) {
	t.Parallel()

	executions := 0
	subject := action.Do(func(context.Context) (int, error) {
		executions++
		return answer, nil
	})

	if executions != 0 {
		t.Fatalf("Action executed during construction: executions=%d", executions)
	}

	value, err := action.Run(t.Context(), subject)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != answer {
		t.Fatalf("Run() value = %d, want %d", value, answer)
	}
	if executions != 1 {
		t.Fatalf("executions = %d, want 1", executions)
	}
}

func TestActionIsReExecutableNotMemoized(t *testing.T) {
	t.Parallel()

	executions := 0
	subject := action.Do(func(context.Context) (int, error) {
		executions++
		return executions, nil
	})

	first, err := action.Run(t.Context(), subject)
	if err != nil {
		t.Fatalf("first Run() error = %v, want nil", err)
	}
	second, err := action.Run(t.Context(), subject)
	if err != nil {
		t.Fatalf("second Run() error = %v, want nil", err)
	}

	if first != 1 || second != 2 {
		t.Fatalf("Run() results = (%d, %d), want (1, 2): Action must not memoize", first, second)
	}
}

func TestActionReceivesTheContextGivenToRun(t *testing.T) {
	t.Parallel()

	subject := action.Do(func(ctx context.Context) (string, error) {
		marker, _ := ctx.Value(contextKey{}).(string)
		return marker, nil
	})

	//nolint:staticcheck // the key is a private struct type, not a bare string.
	ctx := context.WithValue(t.Context(), contextKey{}, "outer")

	value, err := action.Run(ctx, subject)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != "outer" {
		t.Fatalf("Run() value = %q, want %q", value, "outer")
	}
}

func TestActionObservesCancellation(t *testing.T) {
	t.Parallel()

	blocked := make(chan struct{})
	subject := action.Do(func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-blocked:
			return answer, nil
		}
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := action.Run(ctx, subject); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

func TestActionObservesDeadline(t *testing.T) {
	t.Parallel()

	blocked := make(chan struct{})
	subject := action.Do(func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-blocked:
			return answer, nil
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()

	if _, err := action.Run(ctx, subject); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestRunPanicsOnNilAction(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), action.Action[int](nil))
	})
}

func assertPanics(t *testing.T, want string, call func()) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("call did not panic, want panic %q", want)
		}
		message, ok := got.(string)
		if !ok || message != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()

	call()
}
