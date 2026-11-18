// Package functional is an umbrella module for small, dependency-free
// functional-programming building blocks; this file is the entire root
// package, which exports nothing.
//
// Public packages are grouped by role under path prefixes (there is no root
// facade and no Go package at the group directories themselves):
//
//   - primitives/function, primitives/predicate, primitives/order,
//     primitives/fold — combinators over ordinary Go values and slices
//     (fold imports exactly values/optional and values/nonempty).
//   - structural/tuple, structural/lazy, structural/iterator,
//     structural/optics — carriers, deferred work, iter.Seq adapters, and
//     typed Lens (iterator imports exactly structural/tuple).
//   - values/optional, values/result, values/either, values/nonempty,
//     values/validated — algebraic value types (validated imports only
//     values/nonempty).
//   - immutable and state — immutable collections and pure plan/Version
//     evolution over T (mesh-isolated from each other in production).
//
// The module has zero dependencies outside the standard library, permanently.
// Public sibling imports are restricted to the values/validated ->
// values/nonempty edge, the fold -> values/optional and fold ->
// values/nonempty edges, and the iterator -> tuple edge. Shared
// internals include functional/internal/eq, functional/internal/core, and
// functional/internal/engine (used by immutable). Production state does not
// import core or engines.
package functional
