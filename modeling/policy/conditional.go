package policy

import (
	"github.com/alvadorncorp/gopackages/functional/primitives/predicate"
	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

const (
	panicNilPredicate   = "modeling/policy: nil predicate"
	panicNilConditional = "modeling/policy: nil conditional"
)

// When makes p applicable only when condition accepts the input.
func When[I, O any](
	condition predicate.Predicate[I],
	p Policy[I, O],
) func(I) optional.Optional[O] {
	if condition == nil {
		panic(panicNilPredicate)
	}
	requirePolicy(p)
	return func(input I) optional.Optional[O] {
		if !condition(input) {
			return optional.None[O]()
		}
		return optional.Some(p(input))
	}
}

// Choose returns a total Policy that invokes exactly one branch.
func Choose[I, O any](
	condition predicate.Predicate[I],
	whenTrue Policy[I, O],
	whenFalse Policy[I, O],
) Policy[I, O] {
	if condition == nil {
		panic(panicNilPredicate)
	}
	requirePolicy(whenTrue)
	requirePolicy(whenFalse)
	return func(input I) O {
		if condition(input) {
			return whenTrue(input)
		}
		return whenFalse(input)
	}
}

// OrElse returns the first applicable result from two conditional policies.
func OrElse[I, O any](
	left func(I) optional.Optional[O],
	right func(I) optional.Optional[O],
) func(I) optional.Optional[O] {
	return FirstApplicable(left, right)
}

// FirstApplicable returns the first Some result in declaration order.
func FirstApplicable[I, O any](
	first func(I) optional.Optional[O],
	rest ...func(I) optional.Optional[O],
) func(I) optional.Optional[O] {
	requireConditional(first)
	for _, candidate := range rest {
		requireConditional(candidate)
	}
	return func(input I) optional.Optional[O] {
		if result := first(input); result.IsSome() {
			return result
		}
		for _, candidate := range rest {
			if result := candidate(input); result.IsSome() {
				return result
			}
		}
		return optional.None[O]()
	}
}

// Fallback turns a conditional policy into a total Policy.
func Fallback[I, O any](
	conditional func(I) optional.Optional[O],
	fallback Policy[I, O],
) Policy[I, O] {
	requireConditional(conditional)
	requirePolicy(fallback)
	return func(input I) O {
		if output, ok := conditional(input).Get(); ok {
			return output
		}
		return fallback(input)
	}
}

func requireConditional[I, O any](conditional func(I) optional.Optional[O]) {
	if conditional == nil {
		panic(panicNilConditional)
	}
}
