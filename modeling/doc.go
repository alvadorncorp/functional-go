// Package modeling provides explicit, compile-time-safe primitives for
// describing domain decisions. Applications own their concrete domain types,
// rules, policies, commands, events, entities, and aggregates.
//
// The package has no runtime, registry, reflection, dispatcher, persistence,
// transport, or infrastructure responsibility. It may use public functional
// capabilities internally, but its exported signatures remain ordinary Go and
// standard-library types.
package modeling
