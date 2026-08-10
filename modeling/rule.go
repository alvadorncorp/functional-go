package modeling

const panicNilRule = "modeling: nil rule"

// Rule checks whether a value satisfies a domain condition.
type Rule[T any] func(T) error

// AllRules composes rules in argument order and returns the first failure.
func AllRules[T any](rules ...Rule[T]) Rule[T] {
	requireRules(rules)
	return func(value T) error {
		for _, rule := range rules {
			if err := rule(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// AnyRule returns success from the first satisfied rule, or the last failure.
func AnyRule[T any](first Rule[T], rest ...Rule[T]) Rule[T] {
	if first == nil {
		panic(panicNilRule)
	}
	requireRules(rest)
	return func(value T) error {
		lastErr := first(value)
		if lastErr == nil {
			return nil
		}
		for _, rule := range rest {
			lastErr = rule(value)
			if lastErr == nil {
				return nil
			}
		}
		return lastErr
	}
}

func requireRules[T any](rules []Rule[T]) {
	for _, rule := range rules {
		if rule == nil {
			panic(panicNilRule)
		}
	}
}
