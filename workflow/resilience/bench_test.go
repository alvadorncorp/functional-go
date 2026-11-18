package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/resilience"
)

var (
	resilienceValueBenchSink int
	resilienceErrorBenchSink error
)

func BenchmarkRetrySuccess(b *testing.B) {
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 3,
		When:     func(error) bool { return true },
	}, action.Pure(1))

	b.ReportAllocs()
	for b.Loop() {
		var err error
		resilienceValueBenchSink, err = action.Run(context.Background(), program)
		if err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}

func BenchmarkRetryImmediate(b *testing.B) {
	transient := errors.New("transient")
	program := resilience.Retry(resilience.RetryPolicy{
		Attempts: 2,
		When:     func(err error) bool { return errors.Is(err, transient) },
		Backoff:  func(int) time.Duration { return 0 },
	}, failingThenSucceeding(transient, 1))

	b.ReportAllocs()
	for b.Loop() {
		var err error
		resilienceValueBenchSink, err = action.Run(context.Background(), program)
		if err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}

func failingThenSucceeding(transient error, value int) action.Action[int] {
	var attempts int
	return func(context.Context) (int, error) {
		attempts++
		if attempts%2 == 1 {
			return 0, transient
		}
		return value, nil
	}
}

func BenchmarkTimeout(b *testing.B) {
	program := resilience.Timeout(time.Second, action.Pure(1))

	b.ReportAllocs()
	for b.Loop() {
		var err error
		resilienceValueBenchSink, err = action.Run(context.Background(), program)
		if err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}
