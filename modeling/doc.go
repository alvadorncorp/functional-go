// Package modeling provides explicit, compile-time-safe primitives for
// describing domain decisions. Applications own their concrete domain types,
// rules, policies, commands, events, entities, and aggregates.
//
// Decision composition uses MapState, MapEvents, Emit, AppendEvents, and
// Combine. Decision remains the canonical root-package type; persistent
// mechanics live behind a Go internal package and are not consumer surface.
//
// The package has no runtime, registry, reflection, dispatcher, persistence,
// transport, or infrastructure responsibility. It may use public functional
// capabilities internally, but its exported signatures remain ordinary Go and
// standard-library types.
package modeling
