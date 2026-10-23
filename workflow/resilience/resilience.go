// Package resilience provides explicit, lazy transformations over action.Action.
// It performs no work until the returned Action is run and starts no goroutine.
package resilience

import "time"

// Backoff returns the delay after the 1-based attempt that just failed.
type Backoff func(attempt int) time.Duration

// RetryPolicy defines selective retry. Attempts includes the first call.
// A nil Backoff retries immediately without waiting.
type RetryPolicy struct {
	Attempts int
	When     func(error) bool
	Backoff  Backoff
}
