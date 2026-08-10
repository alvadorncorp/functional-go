// Package modeling provides explicit, compile-time-safe primitives for
// describing domain decisions. Applications own their concrete domain types,
// errors, policies, commands, events, entities, and aggregates.
//
// Rule checks one value; Requirement checks state plus operation input;
// Transition computes the next state; Invariant validates the result. The
// application invokes those functions explicitly before constructing a
// Decision. AllRules and AllRequirements are ordered fail-fast composition;
// AnyRule and AnyRequirement are non-empty ordered first-success composition.
//
// Decision composition uses MapState, MapEvents, Emit, AppendEvents, and
// Combine. Decision remains the canonical root-package type; persistent
// mechanics live behind a Go internal package and are not consumer surface.
//
// The package has no engine, runner, registry, reflection, dispatcher,
// persistence, transport, or infrastructure responsibility. Exported
// signatures remain ordinary Go and standard-library types.
package modeling
