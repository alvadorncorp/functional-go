package workflow

import "github.com/alvadorncorp/gopackages/workflow/action"

// Map transforms the output of wf with a pure function, leaving failures
// untouched.
//
// Nothing runs until the resulting Workflow is applied and its Action is run.
// When wf fails, mapper is not called and the error is returned unchanged. Map
// panics with "workflow: nil workflow" or "workflow: nil mapper" when its
// arguments are nil.
func Map[I, A, B any](wf Workflow[I, A], mapper func(A) B) Workflow[I, B] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input I) action.Action[B] {
		return action.Map(apply(wf, input), mapper)
	}
}

// MapFallible transforms the output of wf with a failing pure function. It is
// the seam where a pure domain function shaped func(A) (B, error) — a decision
// function, a validation, a translation between bounded contexts — joins a
// workflow without being turned into an Action.
//
// When wf fails, mapper is not called; when mapper fails, its error is returned
// unchanged. MapFallible panics with "workflow: nil workflow" or
// "workflow: nil mapper" when its arguments are nil.
func MapFallible[I, A, B any](wf Workflow[I, A], mapper func(A) (B, error)) Workflow[I, B] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input I) action.Action[B] {
		return action.MapFallible(apply(wf, input), mapper)
	}
}

// Then composes two workflows sequentially: Workflow[A,B] followed by
// Workflow[B,C] yields Workflow[A,C], and the result is an ordinary Workflow
// that larger compositions cannot distinguish from a primitive one.
//
// Execution order is deterministic — first, then second — and second is not
// applied when first fails. Composition is binary rather than variadic because
// Go generics cannot preserve a changing intermediate type across a variadic
// list; chain calls instead (Then(Then(a, b), c)). Then panics with
// "workflow: nil workflow" when either argument is nil.
func Then[A, B, C any](first Workflow[A, B], second Workflow[B, C]) Workflow[A, C] {
	if first == nil || second == nil {
		panic(panicNilWorkflow)
	}
	return func(input A) action.Action[C] {
		return action.FlatMap(apply(first, input), func(intermediate B) action.Action[C] {
			return apply(second, intermediate)
		})
	}
}
