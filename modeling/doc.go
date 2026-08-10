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
// Pure S-to-S state evolution lives in the separate public subpackage
// github.com/alvadorncorp/gopackages/modeling/change. The root package does not
// re-export Change or manage its execution.
//
// The policy subpackage derives pure domain results from an explicit context.
// It owns transformation and Optional-based conditional selection without
// adding Policy aliases to this root package.
//
// Decision composition uses MapState, MapEvents, Emit, AppendEvents, and
// Combine. Decision remains the canonical root-package type; persistent
// mechanics live behind a Go internal package and are not consumer surface.
//
// Fact-to-state evolution lives in the separate modeling/reducer capability.
// It defines Reducer and ordered Replay without moving or changing root types.
//
// Derived-representation evolution lives in the separate modeling/projection
// capability. It defines Projection and ordered Project without root aliases,
// persistence, subscriptions, or CQRS runtime behavior.
//
// The package has no engine, runner, registry, reflection, dispatcher, state
// manager, persistence, transport, or infrastructure responsibility. Exported
// signatures remain ordinary Go and standard-library types.
package modeling
