package action

// Panic messages are stable, unexported constants. Every combinator that
// rejects a nil argument panics eagerly at construction with one of these,
// because a nil action or a nil callback is a programming defect rather than a
// failure the error channel should carry.
const (
	panicNilAction   = "workflow/action: nil action"
	panicNilFunction = "workflow/action: nil function"
	panicNilError    = "workflow/action: nil error"
	panicNilMapper   = "workflow/action: nil mapper"
	panicNilBinder   = "workflow/action: nil binder"
	panicNilRecovery = "workflow/action: nil recovery"
)
