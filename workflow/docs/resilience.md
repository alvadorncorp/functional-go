# Resilience

`workflow/resilience` transforms an `action.Action[T]` without running it. The
policy stays visible in the adapter that knows the budget and the operational
failure class.

## Selective retry

```go
policy := resilience.RetryPolicy{
    Attempts: 3,
    When: faults.IsUnavailable,
    Backoff: exponentialBackoff,
}
program := resilience.Retry(policy, source)
value, err := action.Run(ctx, program)
```

The adapter can cap the backoff before the shift, using `math/bits` from the
standard library:

```go
func exponentialBackoff(attempt int) time.Duration {
    const baseDelay = 100 * time.Millisecond
    const maxDelay = 2 * time.Second
    if attempt <= 1 {
        return baseDelay
    }
    if attempt-1 >= bits.Len64(uint64(maxDelay/baseDelay)) {
        return maxDelay
    }
    return baseDelay << (attempt - 1)
}
```

The cap before the shift prevents overflow even with a high attempt count.

`Attempts` includes the initial call. The Backoff argument is the 1-based
number of the attempt that failed. `When` does not run after the final
failure. A nil Backoff or a zero wait is immediate; a negative wait panics.
Cancellation before an attempt or during the wait returns the exact context
error and the zero value of T.

Retry does not make an operation safe. Use it only when the effect is
idempotent, deduplicated, transactional at the required boundary, or when
partial repetition has been modeled. The classification predicate belongs to
the adapter, not the package.

## Cooperative timeout

```go
program := resilience.Timeout(3*time.Second, source)
value, err := action.Run(request.Context(), program)
```

Timeout derives a child context and calls the source on the same goroutine. It
is cooperative: if the source ignores the context, there is no magic
preemption.

## Composition order

```go
totalBudget := resilience.Timeout(30*time.Second,
    resilience.Retry(policy, source),
)

perAttemptPolicy := policy
perAttemptPolicy.When = func(err error) bool {
    return faults.IsUnavailable(err) || errors.Is(err, context.DeadlineExceeded)
}
perAttemptBudget := resilience.Retry(perAttemptPolicy,
    resilience.Timeout(3*time.Second, source),
)
```

`Timeout(Retry(source))` bounds the whole set of attempts and backoffs.
`Retry(Timeout(source))` renews the budget for each attempt, which is why the
second predicate also accepts `context.DeadlineExceeded` from the child
context. Retry's `ctx.Err()` check before each attempt and its cancelable wait
keep respecting the parent context: once its total budget runs out, there is
no new attempt. The timeout stays cooperative even if the source ignores the
context.

## Transparency

Construction is lazy. The package starts no goroutine, recovers no panic,
wraps no errors, records no metrics, and stores no global state. Plan
describes structure; Resilience transforms Actions; only Run executes.
