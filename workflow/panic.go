package workflow

// Panic messages are stable, unexported constants. A nil workflow or a nil
// callback is a programming defect, so every combinator rejects one eagerly at
// construction rather than deferring the failure to execution.
const (
	panicNilWorkflow = "workflow: nil workflow"
	panicNilAction   = "workflow: nil action"
	panicNilFunction = "workflow: nil function"
	panicNilMapper   = "workflow: nil mapper"
)
