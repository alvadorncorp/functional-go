package resilience_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvadorncorp/gopackages/workflow/action"
	"github.com/alvadorncorp/gopackages/workflow/resilience"
)

func TestRetryPanicsOnInvalidAttempts(t *testing.T) {
	t.Parallel()

	var sourceCalls, whenCalls atomic.Int32
	source := countingSource(&sourceCalls, 1, nil)
	when := countingWhen(&whenCalls, true)

	for _, attempts := range []int{0, -1} {
		policy := resilience.RetryPolicy{Attempts: attempts, When: when}
		assertPanic(t, "workflow/resilience: invalid attempts", func() {
			resilience.Retry(policy, source)
		})
	}
	if sourceCalls.Load() != 0 || whenCalls.Load() != 0 {
		t.Fatalf("source/predicate calls = %d/%d, want 0/0", sourceCalls.Load(), whenCalls.Load())
	}
}

func TestRetryPanicsOnNilPredicate(t *testing.T) {
	t.Parallel()

	assertPanic(t, "workflow/resilience: nil predicate", func() {
		resilience.Retry(resilience.RetryPolicy{Attempts: 1}, action.Pure(1))
	})
}

func TestRetryPanicsOnNilSource(t *testing.T) {
	t.Parallel()

	assertPanic(t, "workflow/resilience: nil action", func() {
		resilience.Retry(resilience.RetryPolicy{
			Attempts: 1,
			When:     func(error) bool { return true },
		}, action.Action[int](nil))
	})
}

func TestRetryIsLazyAtConstruction(t *testing.T) {
	t.Parallel()

	var sourceCalls atomic.Int32
	source := countingSource(&sourceCalls, 1, nil)
	_ = resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     func(error) bool { return true },
	}, source)
	if sourceCalls.Load() != 0 {
		t.Fatalf("source calls at construction = %d, want 0", sourceCalls.Load())
	}
}

func TestRetryReturnsOnFirstSuccess(t *testing.T) {
	t.Parallel()

	var sourceCalls, whenCalls, backoffCalls atomic.Int32
	source := countingSource(&sourceCalls, 42, nil)
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     countingWhen(&whenCalls, true),
		Backoff:  countingBackoff(&backoffCalls, 0),
	}, source)

	value, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 42 {
		t.Fatalf("Run() value = %d, want 42", value)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1", sourceCalls.Load())
	}
	if whenCalls.Load() != 0 || backoffCalls.Load() != 0 {
		t.Fatalf("when/backoff calls = %d/%d, want 0/0", whenCalls.Load(), backoffCalls.Load())
	}
}

func TestRetryThenSuccessRecordsAttempts(t *testing.T) {
	t.Parallel()

	transient := errors.New("transient")
	var (
		sourceCalls  atomic.Int32
		whenCalls    atomic.Int32
		backoffArgs  []int
		backoffMu    sync.Mutex
		sourceErrSeq = []error{transient, transient, nil}
	)
	source := action.Do(func(context.Context) (int, error) {
		n := int(sourceCalls.Add(1))
		if n > len(sourceErrSeq) {
			t.Fatalf("unexpected source call %d", n)
		}
		err := sourceErrSeq[n-1]
		if err != nil {
			return 0, err
		}
		return 7, nil
	})
	policy := resilience.RetryPolicy{
		Attempts: 3,
		When:     countingWhen(&whenCalls, true),
		Backoff: func(attempt int) time.Duration {
			backoffMu.Lock()
			backoffArgs = append(backoffArgs, attempt)
			backoffMu.Unlock()
			return 0
		},
	}

	value, err := action.Run(t.Context(), resilience.Retry(policy, source))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 7 {
		t.Fatalf("Run() value = %d, want 7", value)
	}
	if sourceCalls.Load() != 3 {
		t.Fatalf("source calls = %d, want 3", sourceCalls.Load())
	}
	if whenCalls.Load() != 2 {
		t.Fatalf("when calls = %d, want 2", whenCalls.Load())
	}
	backoffMu.Lock()
	gotBackoff := append([]int(nil), backoffArgs...)
	backoffMu.Unlock()
	if len(gotBackoff) != 2 || gotBackoff[0] != 1 || gotBackoff[1] != 2 {
		t.Fatalf("backoff args = %v, want [1 2]", gotBackoff)
	}
}

func TestRetryExhaustedPreservesErrorIdentity(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("exhausted")
	var sourceCalls, whenCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		sourceCalls.Add(1)
		return 0, sentinel
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     countingWhen(&whenCalls, true),
		Backoff:  func(int) time.Duration { return 0 },
	}, source)

	value, err := action.Run(t.Context(), program)
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v, want sentinel", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 3 {
		t.Fatalf("source calls = %d, want 3", sourceCalls.Load())
	}
	if whenCalls.Load() != 2 {
		t.Fatalf("when calls = %d, want 2 (Attempts-1)", whenCalls.Load())
	}
}

func TestRetryStopsWhenPredicateFalse(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("not retryable")
	var sourceCalls, whenCalls, backoffCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		sourceCalls.Add(1)
		return 0, sentinel
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 5,
		When:     countingWhen(&whenCalls, false),
		Backoff:  countingBackoff(&backoffCalls, 0),
	}, source)

	value, err := action.Run(t.Context(), program)
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v, want sentinel", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1", sourceCalls.Load())
	}
	if whenCalls.Load() != 1 {
		t.Fatalf("when calls = %d, want 1", whenCalls.Load())
	}
	if backoffCalls.Load() != 0 {
		t.Fatalf("backoff calls = %d, want 0", backoffCalls.Load())
	}
}

func TestRetryPreservesFinalDistinctError(t *testing.T) {
	t.Parallel()

	first := errors.New("first")
	second := errors.New("second")
	third := errors.New("third")
	seq := []error{first, second, third}
	var sourceCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		n := int(sourceCalls.Add(1))
		return 0, seq[n-1]
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return 0 },
	}, source)

	value, err := action.Run(t.Context(), program)
	if !errors.Is(err, third) {
		t.Fatalf("Run() error = %v, want third", err)
	}
	if err == first || err == second {
		t.Fatalf("Run() error identity = %v, want third only", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
}

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

func TestRetryNegativeBackoffPanicsAfterPredicate(t *testing.T) {
	t.Parallel()

	transient := errors.New("transient")
	var sourceCalls, whenCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		sourceCalls.Add(1)
		return 0, transient
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     countingWhen(&whenCalls, true),
		Backoff:  func(int) time.Duration { return -1 },
	}, source)

	assertPanic(t, "workflow/resilience: negative backoff", func() {
		_, _ = action.Run(t.Context(), program)
	})
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1", sourceCalls.Load())
	}
	if whenCalls.Load() != 1 {
		t.Fatalf("when calls = %d, want 1", whenCalls.Load())
	}
}

func TestRetryPreCanceledContext(t *testing.T) {
	t.Parallel()

	var sourceCalls, whenCalls, backoffCalls atomic.Int32
	source := countingSource(&sourceCalls, 1, nil)
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     countingWhen(&whenCalls, true),
		Backoff:  countingBackoff(&backoffCalls, time.Second),
	}, source)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	value, err := action.Run(ctx, program)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 0 || whenCalls.Load() != 0 || backoffCalls.Load() != 0 {
		t.Fatalf("calls source/when/backoff = %d/%d/%d, want 0/0/0",
			sourceCalls.Load(), whenCalls.Load(), backoffCalls.Load())
	}
}

func TestRetryCancellationDuringBackoff(t *testing.T) {
	t.Parallel()

	transient := errors.New("transient")
	var sourceCalls atomic.Int32
	started := make(chan struct{})
	source := action.Do(func(context.Context) (int, error) {
		sourceCalls.Add(1)
		return 0, transient
	})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     func(error) bool { return true },
		Backoff: func(int) time.Duration {
			close(started)
			return time.Hour
		},
	}, source)

	done := make(chan struct{})
	var (
		value int
		err   error
	)
	go func() {
		defer close(done)
		value, err = action.Run(ctx, program)
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("backoff did not start")
	}
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not complete after cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1", sourceCalls.Load())
	}
}

func TestRetryDeadlineDuringBackoff(t *testing.T) {
	t.Parallel()

	transient := errors.New("transient")
	var sourceCalls atomic.Int32
	source := action.Do(func(context.Context) (int, error) {
		sourceCalls.Add(1)
		return 0, transient
	})
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return time.Hour },
	}, source)

	value, err := action.Run(ctx, program)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context.DeadlineExceeded", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1", sourceCalls.Load())
	}
}

func TestRetrySourcePanicEscapes(t *testing.T) {
	t.Parallel()

	sentinel := "source panic"
	source := action.Do(func(context.Context) (int, error) {
		panic(sentinel)
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
	}, source)

	assertPanic(t, sentinel, func() {
		_, _ = action.Run(t.Context(), program)
	})
}

func TestRetryPredicatePanicEscapes(t *testing.T) {
	t.Parallel()

	sentinel := "predicate panic"
	source := action.Do(func(context.Context) (int, error) {
		return 0, errors.New("fail")
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { panic(sentinel) },
	}, source)

	assertPanic(t, sentinel, func() {
		_, _ = action.Run(t.Context(), program)
	})
}

func TestRetryBackoffPanicEscapes(t *testing.T) {
	t.Parallel()

	sentinel := "backoff panic"
	source := action.Do(func(context.Context) (int, error) {
		return 0, errors.New("fail")
	})
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { panic(sentinel) },
	}, source)

	assertPanic(t, sentinel, func() {
		_, _ = action.Run(t.Context(), program)
	})
}

func TestRetryConcurrentReuse(t *testing.T) {
	t.Parallel()

	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(error) bool { return true },
		Backoff:  func(int) time.Duration { return 0 },
	}, action.Pure(11))

	const goroutines = 32
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			value, err := action.Run(context.Background(), program)
			if err != nil {
				errs <- err
				return
			}
			if value != 11 {
				errs <- fmt.Errorf("value = %d, want 11", value)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func countingSource(calls *atomic.Int32, value int, err error) action.Action[int] {
	return action.Do(func(context.Context) (int, error) {
		calls.Add(1)
		return value, err
	})
}

func countingWhen(calls *atomic.Int32, retryable bool) func(error) bool {
	return func(error) bool {
		calls.Add(1)
		return retryable
	}
}

func countingBackoff(calls *atomic.Int32, delay time.Duration) resilience.Backoff {
	return func(int) time.Duration {
		calls.Add(1)
		return delay
	}
}

func assertPanic(t *testing.T, want any, fn func()) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("call did not panic, want %v", want)
		}
		if got != want {
			t.Fatalf("panic = %#v, want %#v", got, want)
		}
	}()
	fn()
}
