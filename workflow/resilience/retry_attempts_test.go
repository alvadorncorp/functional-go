package resilience_test

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/resilience"
)

func TestRetryZeroBackoffIsImmediate(t *testing.T) {
	t.Parallel()

	transient := errors.New("transient")
	var sourceCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		n := sourceCalls.Add(1)
		if n == 1 {
			return 0, transient
		}
		return 9, nil
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return 0 },
	}, source)

	value, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 9 || sourceCalls.Load() != 2 {
		t.Fatalf("value/calls = %d/%d, want 9/2", value, sourceCalls.Load())
	}
}

func TestRetryNilBackoffIsImmediate(t *testing.T) {
	t.Parallel()

	const attempts = 3
	assertRetryAttemptCount(t, attempts, attempts)
}

func TestRetryPositiveBackoffThenSuccess(t *testing.T) {
	t.Parallel()

	const attempts = 3
	transient := errors.New("transient")
	calls := 0
	source := action.Do(func(context.Context) (int, error) {
		calls++
		if calls < attempts {
			return 0, transient
		}
		return 7, nil
	})
	var backoffArgs []int
	policy := resilience.RetryPolicy{
		Attempts: attempts,
		When:     func(err error) bool { return err == transient },
		Backoff: func(attempt int) time.Duration {
			backoffArgs = append(backoffArgs, attempt)
			return time.Millisecond
		},
	}

	value, err := action.Run(t.Context(), resilience.Retry(policy, source))
	if err != nil || value != 7 || calls != attempts {
		t.Fatalf("Run() = (%d, %v), calls = %d; want (7, nil), calls = 3", value, err, calls)
	}
	if !slices.Equal(backoffArgs, []int{1, 2}) {
		t.Fatalf("backoff attempts = %v, want [1 2]", backoffArgs)
	}
}

func TestRetryAttemptCounts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		attempts  int
		successAt int
	}{
		{name: "first-success", attempts: 1, successAt: 1},
		{name: "single-failure", attempts: 1, successAt: 3},
		{name: "early-success", attempts: 3, successAt: 1},
		{name: "middle-success", attempts: 3, successAt: 2},
		{name: "last-success", attempts: 3, successAt: 3},
		{name: "exhausted", attempts: 3, successAt: 5},
		{name: "remaining-attempts-unused", attempts: 5, successAt: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertRetryAttemptCount(t, tc.attempts, tc.successAt)
		})
	}
}

func FuzzRetryAttempts(f *testing.F) {
	for _, seed := range []struct{ attempts, successAt uint8 }{
		{0, 0}, {0, 2}, {2, 0}, {2, 1}, {2, 2}, {2, 4}, {4, 2}, {255, 255},
	} {
		f.Add(seed.attempts, seed.successAt)
	}
	f.Fuzz(func(t *testing.T, attempts, successAt uint8) {
		assertRetryAttemptCount(t, int(attempts)+1, int(successAt)+1)
	})
}

func assertRetryAttemptCount(t *testing.T, attempts, successAt int) {
	t.Helper()

	transient := errors.New("transient")
	calls, whenCalls := 0, 0
	source := action.Do(func(context.Context) (int, error) {
		calls++
		if calls == successAt {
			return calls, nil
		}
		return -1, transient
	})
	policy := resilience.RetryPolicy{
		Attempts: attempts,
		When: func(err error) bool {
			whenCalls++
			if err != transient {
				t.Errorf("When error = %v, want exact transient error", err)
			}
			return true
		},
	}

	value, err := action.Run(t.Context(), resilience.Retry(policy, source))
	wantCalls := min(attempts, successAt)
	if calls != wantCalls || whenCalls != wantCalls-1 {
		t.Fatalf("attempts=%d successAt=%d: source/predicate calls = %d/%d, want %d/%d",
			attempts, successAt, calls, whenCalls, wantCalls, wantCalls-1)
	}
	if successAt <= attempts {
		if err != nil || value != successAt {
			t.Fatalf("Run() = (%d, %v), want (%d, nil)", value, err, successAt)
		}
	} else if err != transient || value != 0 {
		t.Fatalf("Run() = (%d, %v), want (0, exact transient error)", value, err)
	}
}
