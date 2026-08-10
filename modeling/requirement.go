package modeling

const panicNilRequirement = "modeling: nil requirement"

// Requirement checks whether an input may be applied to a state.
type Requirement[S, I any] func(S, I) error

// AllRequirements composes requirements in argument order and returns the first failure.
func AllRequirements[S, I any](requirements ...Requirement[S, I]) Requirement[S, I] {
	requireRequirements(requirements)
	return func(state S, input I) error {
		for _, requirement := range requirements {
			if err := requirement(state, input); err != nil {
				return err
			}
		}
		return nil
	}
}

// AnyRequirement returns success from the first satisfied requirement, or the last failure.
func AnyRequirement[S, I any](first Requirement[S, I], rest ...Requirement[S, I]) Requirement[S, I] {
	if first == nil {
		panic(panicNilRequirement)
	}
	requireRequirements(rest)
	return func(state S, input I) error {
		lastErr := first(state, input)
		if lastErr == nil {
			return nil
		}
		for _, requirement := range rest {
			lastErr = requirement(state, input)
			if lastErr == nil {
				return nil
			}
		}
		return lastErr
	}
}

func requireRequirements[S, I any](requirements []Requirement[S, I]) {
	for _, requirement := range requirements {
		if requirement == nil {
			panic(panicNilRequirement)
		}
	}
}
