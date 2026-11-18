package modeling

import internaldecision "go.alvadorn.org/functional/modeling/internal/decision"

// MapState transforms a Decision state while preserving its events.
func MapState[S, T, E any](decision Decision[S, E], mapper func(S) T) Decision[T, E] {
	return Decision[T, E]{value: internaldecision.MapState(decision.value, mapper)}
}

// MapEvents transforms a Decision's events in order while preserving its state.
func MapEvents[S, E, F any](decision Decision[S, E], mapper func(E) F) Decision[S, F] {
	return Decision[S, F]{value: internaldecision.MapEvents(decision.value, mapper)}
}

// Emit returns a new Decision with event appended.
func Emit[S, E any](decision Decision[S, E], event E) Decision[S, E] {
	return AppendEvents(decision, event)
}

// AppendEvents returns a new Decision with events appended in argument order.
func AppendEvents[S, E any](decision Decision[S, E], events ...E) Decision[S, E] {
	decision.value = decision.value.Append(events...)
	return decision
}

// Combine returns the right state with left events followed by right events.
func Combine[S, E any](left, right Decision[S, E]) Decision[S, E] {
	return Decision[S, E]{value: internaldecision.Combine(left.value, right.value)}
}
