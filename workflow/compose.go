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

// Contramap adapts the input of wf, so one workflow serves several callers that
// hold different input shapes — an HTTP DTO, a CLI argument set, a consumed
// message — without any of them changing the workflow.
//
// mapper runs when the resulting Workflow is applied, before any effect.
// Contramap panics with "workflow: nil workflow" or "workflow: nil mapper" when
// its arguments are nil.
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

// FlatMap sequences a dependent Action after wf, for the case where the next
// effect is chosen from the previous output — a domain policy selecting a
// branch, for instance.
//
// binder runs only on success, and the Action it returns runs with the same
// context. FlatMap panics with "workflow: nil workflow" or
// "workflow: nil binder" at construction, and with "workflow: nil action" at
// execution when binder returns nil.
func FlatMap[I, A, B any](wf Workflow[I, A], binder func(A) action.Action[B]) Workflow[I, B] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if binder == nil {
		panic(panicNilBinder)
	}
	return func(input I) action.Action[B] {
		return action.FlatMap(apply(wf, input), binder)
	}
}

// Tap runs a side effect over the output of wf and yields that output
// unchanged.
//
// The observed Action runs after wf and before the value is returned, with the
// same context; a failure of either propagates unchanged. Tap panics with
// "workflow: nil workflow" or "workflow: nil observer" at construction, and
// with "workflow: nil action" at execution when observer returns nil.
func Tap[I, O any](wf Workflow[I, O], observer func(O) action.Action[struct{}]) Workflow[I, O] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if observer == nil {
		panic(panicNilObserver)
	}
	return func(input I) action.Action[O] {
		return action.Tap(apply(wf, input), observer)
	}
}

// Recover replaces a failure of wf with the Action recovery returns.
//
// recovery receives the error unchanged, including context.Canceled and
// context.DeadlineExceeded, which this module never reclassifies. Use it for
// genuine failure handling; an expected branch belongs in the domain as a
// value, not in the error channel. Recover panics with
// "workflow: nil workflow" or "workflow: nil recovery" at construction, and
// with "workflow: nil action" at execution when recovery returns nil.
func Recover[I, O any](wf Workflow[I, O], recovery func(error) action.Action[O]) Workflow[I, O] {
	if wf == nil {
		panic(panicNilWorkflow)
	}
	if recovery == nil {
		panic(panicNilRecovery)
	}
	return func(input I) action.Action[O] {
		return action.Recover(apply(wf, input), recovery)
	}
}
