package workflow

import (
	"github.com/alvadorncorp/gopackages/workflow/action"
)

// Workflow is a reusable operation that turns an input into a deferred
// computation. Dependencies are captured when the workflow is constructed, the
// input arrives when the workflow is applied, and a context.Context arrives
// only when the resulting action.Action is run.
//
// Applying a Workflow performs no effect: it returns a description of the work.
// A Workflow holds no execution state, so one constructed value serves any
// number of inputs, contexts, and goroutines concurrently, and is exactly as
// safe for concurrent use as the dependencies its closure captured.
//
// A Workflow is not a handler, controller, service object, job, durable
// workflow, or state machine, and it carries no persisted lifecycle of its own,
// though any of those boundaries may execute one.
//
// The zero value is a nil function; applying it panics. Every combinator in
// this package rejects a nil Workflow at construction with
// "workflow: nil workflow".
type Workflow[I, O any] func(I) action.Action[O]

// From builds a Workflow from a function that maps an input to an Action. It is
// the idiomatic bridge from a capability — a small struct of functions supplied
// by the composition root — to a composable workflow.
//
// fn is not called by From. From panics with "workflow: nil function" when fn
// is nil.
func From[I, O any](fn func(I) action.Action[O]) Workflow[I, O] {
	if fn == nil {
		panic(panicNilFunction)
	}
	return fn
}

// apply applies wf to input and rejects a Workflow that yields a nil Action,
// so the failure names the workflow rather than surfacing as an opaque nil
// function panic deep inside execution.
func apply[I, O any](wf Workflow[I, O], input I) action.Action[O] {
	result := wf(input)
	if result == nil {
		panic(panicNilAction)
	}
	return result
}
