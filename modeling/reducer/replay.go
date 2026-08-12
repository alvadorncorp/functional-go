package reducer

import "iter"

// Replay evolves initial state by applying reduce once to each event in
// encounter order. An empty sequence returns initial unchanged.
//
// Replay is a single-pass left fold: it consumes events exactly once and
// never materializes []E or an intermediate collection. Callers that need
// multi-pass access must materialize explicitly before invoking Replay.
//
// Replay has no knowledge of event storage, streams, versions, serialization,
// publication, or any other Event Sourcing runtime concern.
func Replay[S, E any](
	initial S,
	events iter.Seq[E],
	reduce Reducer[S, E],
) S {
	state := initial
	for event := range events {
		state = reduce(state, event)
	}
	return state
}
