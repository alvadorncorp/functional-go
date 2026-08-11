package workflow

import "github.com/alvadorncorp/gopackages/workflow/action"

// Provide fixes the input of wf and returns its deferred Action. It applies the
// workflow immediately but does not run the Action. Provide panics with
// "workflow: nil workflow" when wf is nil and with "workflow: nil action" when
// wf returns nil.
func Provide[I, O any](wf Workflow[I, O], input I) action.Action[O] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	return apply(wf, input)
}

// Contramap adapts a caller input A to the input I understood by wf. mapper is
// called only when the returned Workflow is applied; the Action produced by wf
// remains deferred until Run. Contramap panics with "workflow: nil workflow"
// or "workflow: nil mapper" when an argument is nil.
func Contramap[A, I, O any](wf Workflow[I, O], mapper func(A) I) Workflow[A, O] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input A) action.Action[O] {
		return apply(wf, mapper(input))
	}
}
