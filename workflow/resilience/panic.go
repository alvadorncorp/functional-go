package resilience

const (
	panicInvalidAttempts = "workflow/resilience: invalid attempts"
	panicNilPredicate    = "workflow/resilience: nil predicate"
	panicNilAction       = "workflow/resilience: nil action"
	panicNegativeBackoff = "workflow/resilience: negative backoff"
	panicInvalidTimeout  = "workflow/resilience: invalid timeout"
)
