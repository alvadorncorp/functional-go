package projection

import "github.com/alvadorncorp/gopackages/functional/immutable"

// Project derives a representation by applying project once to every event in
// encounter order. An empty event list returns initial unchanged.
//
// Project is the batch form of the same incremental P x E -> P operation. It
// loads no events, persists no result, and provides no CQRS runtime behavior.
func Project[P, E any](
	initial P,
	events immutable.List[E],
	project Projection[P, E],
) P {
	projected := initial
	for event := range events.Values() {
		projected = project(projected, event)
	}
	return projected
}
