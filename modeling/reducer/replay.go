package reducer

import "github.com/alvadorncorp/gopackages/functional/immutable"

// Replay evolves initial state by applying reduce once to each event in
// encounter order. An empty event list returns initial unchanged.
//
// Replay is semantically a left fold over facts that have already occurred.
// It has no knowledge of event storage, streams, versions, serialization,
// publication, or any other Event Sourcing runtime concern.
func Replay[S, E any](
	initial S,
	events immutable.List[E],
	reduce Reducer[S, E],
) S {
	state := initial
	for event := range events.Values() {
		state = reduce(state, event)
	}
	return state
}
