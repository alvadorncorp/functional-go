package modeling

// Transition is a pure domain transition from state and input to output.
// Ordinary functions with the same signature are assignable to it.
type Transition[S, I, O any] func(S, I) O

// Apply invokes t explicitly as a transition.
func (t Transition[S, I, O]) Apply(state S, input I) O {
	return t(state, input)
}

// FallibleTransition is a domain transition that may return a domain failure.
// It uses ordinary Go errors rather than introducing a Result or Either type.
type FallibleTransition[S, I, O any] func(S, I) (O, error)

// Apply invokes t explicitly as a fallible transition.
func (t FallibleTransition[S, I, O]) Apply(state S, input I) (O, error) {
	return t(state, input)
}
