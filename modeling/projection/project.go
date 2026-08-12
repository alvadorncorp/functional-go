package projection

import "iter"

// Project derives a representation by applying project once to every event in
// encounter order. An empty sequence returns initial unchanged.
//
// Project is a single-pass left fold: it consumes events exactly once and
// never materializes []E or an intermediate collection. Callers that need
// multi-pass access must materialize explicitly before invoking Project.
//
// Project is the batch form of the same incremental P x E -> P operation. It
// loads no events, persists no result, and provides no CQRS runtime behavior.
func Project[P, E any](
	initial P,
	events iter.Seq[E],
	project Projection[P, E],
) P {
	projected := initial
	for event := range events {
		projected = project(projected, event)
	}
	return projected
}
