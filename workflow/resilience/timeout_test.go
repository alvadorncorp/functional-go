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

type ctxKey string

const parentValueKey ctxKey = "parent-value"

func TestTimeoutPanicsOnInvalidDuration(t *testing.T) {
	t.Parallel()

	var sourceCalls atomic.Int32
	source := countingSource(&sourceCalls, 1, nil)
	for _, duration := range []time.Duration{0, -time.Nanosecond} {
		assertPanic(t, "workflow/resilience: invalid timeout", func() {
			resilience.Timeout(duration, source)
		})
	}
	if sourceCalls.Load() != 0 {
		t.Fatalf("source calls = %d, want 0", sourceCalls.Load())
	}
}

func TestTimeoutPanicsOnNilSource(t *testing.T) {
	t.Parallel()

	assertPanic(t, "workflow/resilience: nil action", func() {
		resilience.Timeout(time.Second, action.Action[int](nil))
	})
}

func TestTimeoutIsLazyAtConstruction(t *testing.T) {
	t.Parallel()

	var sourceCalls atomic.Int32
	source := countingSource(&sourceCalls, 1, nil)
	_ = resilience.Timeout(time.Second, source)
	if sourceCalls.Load() != 0 {
		t.Fatalf("source calls at construction = %d, want 0", sourceCalls.Load())
	}
}

func TestTimeoutDerivesChildDeadline(t *testing.T) {
	t.Parallel()

	parent := context.Background()
	var (
		sawChild    context.Context
		deadline    time.Time
		hasDeadline bool
	)
	source := action.Do(func(ctx context.Context) (int, error) {
		sawChild = ctx
		deadline, hasDeadline = ctx.Deadline()
		return 1, nil
	})

	before := time.Now()
	value, err := action.Run(parent, resilience.Timeout(time.Second, source))
	after := time.Now()
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 1 {
		t.Fatalf("Run() value = %d, want 1", value)
	}
	if sawChild == nil || sawChild == parent {
		t.Fatalf("source saw parent context, want distinct child")
	}
	if !hasDeadline {
		t.Fatal("child context has no deadline")
	}
	// Deadline must not be later than approximately now+duration (allow small clock skew).
	latest := after.Add(time.Second + 50*time.Millisecond)
	if deadline.After(latest) {
		t.Fatalf("deadline %v after latest acceptable %v (before=%v)", deadline, latest, before)
	}
	earliest := before.Add(500 * time.Millisecond)
	if deadline.Before(earliest) {
		t.Fatalf("deadline %v before earliest acceptable %v", deadline, earliest)
	}
}

func TestTimeoutPropagatesParentValues(t *testing.T) {
	t.Parallel()

	parent := context.WithValue(context.Background(), parentValueKey, "visible")
	var seen any
	source := action.Do(func(ctx context.Context) (int, error) {
		seen = ctx.Value(parentValueKey)
		return 1, nil
	})

	if _, err := action.Run(parent, resilience.Timeout(time.Second, source)); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if seen != "visible" {
		t.Fatalf("parent value = %v, want visible", seen)
	}
}

func TestTimeoutPreservesImmediateResultIdentity(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("immediate")
	source := action.Do(func(context.Context) (int, error) {
		return 5, sentinel
	})

	value, err := action.Run(t.Context(), resilience.Timeout(time.Second, source))
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v, want sentinel", err)
	}
	if value != 5 {
		t.Fatalf("Run() value = %d, want 5", value)
	}
}

func TestTimeoutSourceBlockedOnDoneReturnsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	source := action.Do(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})

	value, err := action.Run(t.Context(), resilience.Timeout(20*time.Millisecond, source))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context.DeadlineExceeded", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
}

func TestTimeoutParentCanceledBeforeRun(t *testing.T) {
	t.Parallel()

	parent, cancel := context.WithCancel(t.Context())
	cancel()

	var seen error
	source := action.Do(func(ctx context.Context) (int, error) {
		seen = ctx.Err()
		return 0, ctx.Err()
	})

	value, err := action.Run(parent, resilience.Timeout(time.Second, source))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if !errors.Is(seen, context.Canceled) {
		t.Fatalf("source observed %v, want context.Canceled", seen)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
}

func TestTimeoutSourcePanicEscapes(t *testing.T) {
	t.Parallel()

	sentinel := "timeout source panic"
	source := action.Do(func(context.Context) (int, error) {
		panic(sentinel)
	})
	assertPanic(t, sentinel, func() {
		_, _ = action.Run(t.Context(), resilience.Timeout(time.Second, source))
	})
}

func TestTimeoutDoesNotPreemptIgnoringSource(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	started := make(chan struct{})
	source := action.Do(func(context.Context) (int, error) {
		close(started)
		<-release
		return 3, nil
	})

	done := make(chan struct{})
	var (
		value int
		err   error
	)
	go func() {
		defer close(done)
		value, err = action.Run(context.Background(), resilience.Timeout(20*time.Millisecond, source))
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("source did not start")
	}

	// Watchdog: prove Run has not returned merely because the deadline elapsed.
	select {
	case <-done:
		t.Fatal("Run returned at deadline despite ignoring context; expected no preemption")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not complete after release")
	}
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 3 {
		t.Fatalf("Run() value = %d, want 3", value)
	}
}

func TestTimeoutConcurrentReuse(t *testing.T) {
	t.Parallel()

	program := resilience.Timeout(time.Second, action.Pure(13))
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
			if value != 13 {
				errs <- fmt.Errorf("value = %d, want 13", value)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestTimeoutOutsideRetryUsesOneTotalBudget(t *testing.T) {
	t.Parallel()

	transient := errors.New("retryable")
	var (
		sourceCalls atomic.Int32
		deadlines   []time.Time
		mu          sync.Mutex
	)
	source := action.Do(func(ctx context.Context) (int, error) {
		sourceCalls.Add(1)
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("source context missing deadline")
		}
		mu.Lock()
		deadlines = append(deadlines, deadline)
		mu.Unlock()
		return 0, transient
	})

	policy := resilience.RetryPolicy{
		Attempts: 5,
		When:     func(err error) bool { return errors.Is(err, transient) },
		Backoff:  func(int) time.Duration { return time.Hour },
	}
	program := resilience.Timeout(40*time.Millisecond, resilience.Retry(policy, source))

	value, err := action.Run(t.Context(), program)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context.DeadlineExceeded", err)
	}
	if value != 0 {
		t.Fatalf("Run() value = %d, want 0", value)
	}
	if sourceCalls.Load() != 1 {
		t.Fatalf("source calls = %d, want 1 (single total budget)", sourceCalls.Load())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(deadlines) != 1 {
		t.Fatalf("recorded deadlines = %d, want 1", len(deadlines))
	}
}

func TestRetryOutsideTimeoutCreatesPerAttemptBudgets(t *testing.T) {
	t.Parallel()

	transient := errors.New("retryable")
	var (
		sourceCalls atomic.Int32
		children    []context.Context
		deadlines   []time.Time
		mu          sync.Mutex
	)
	source := action.Do(func(ctx context.Context) (int, error) {
		n := sourceCalls.Add(1)
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("source context missing deadline")
		}
		mu.Lock()
		children = append(children, ctx)
		deadlines = append(deadlines, deadline)
		mu.Unlock()
		if n == 1 {
			return 0, transient
		}
		return 21, nil
	})

	policy := resilience.RetryPolicy{
		Attempts: 2,
		When:     func(err error) bool { return errors.Is(err, transient) },
		Backoff:  func(int) time.Duration { return 0 },
	}
	program := resilience.Retry(policy, resilience.Timeout(time.Second, source))

	value, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != 21 {
		t.Fatalf("Run() value = %d, want 21", value)
	}
	if sourceCalls.Load() != 2 {
		t.Fatalf("source calls = %d, want 2", sourceCalls.Load())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(children) != 2 || len(deadlines) != 2 {
		t.Fatalf("children/deadlines = %d/%d, want 2/2", len(children), len(deadlines))
	}
	if children[0] == children[1] {
		t.Fatal("both attempts shared the same derived context")
	}
	// Fresh deadlines: second attempt's deadline should not be identical to the first.
	if deadlines[0].Equal(deadlines[1]) {
		t.Fatalf("deadlines equal (%v); want fresh per-attempt budgets", deadlines[0])
	}
}
