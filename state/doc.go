// Package state is a semantic evolution layer for values of type T (any).
//
// It models versioned pure state over T: each evolution yields a new state
// value rather than mutating the receiver in place. The package owns
// lineage, version identity, single-use transactions, and structural diff
// contracts for that pure T. It does not own domain rules, validation, or
// persistence engines.
//
// # Stability
//
// The existing state API is supported and frozen: it keeps working as
// documented, but the package is not extended with new surface. The package
// carries no Deprecated marker, and there is no removal or v2 plan for it.
//
// # Pure T, no domain
//
// State[T] holds and evolves T as opaque data. Callers supply T and any
// domain meaning; this package never interprets fields of T and never
// imports domain packages. Prefer immutable or value-object T when deep
// immutability of the payload is required — the same semantic-immutability
// caveat that applies to immutable collections applies here.
//
// # Evolution surface
//
// Of seeds a Version-0 plan with a fresh lineage token. Apply records a
// pure Mutation (func(T) T) as one reverse-linked plan node without
// invoking it. Materialize re-applies the full plan oldest→newest on every
// call (no cache). Snapshot freezes Version and Value at one materialization
// point. Begin opens a single-use Transaction that batches plan growth;
// Commit returns a State and does not Materialize. DiffBetween captures
// structural plan steps between two States of the same lineage when the
// destination is a descendant of the source.
//
// # Distinctions
//
//	state.Transaction ≠ immutable.Builder
//	state.Diff ≠ Domain Event
//	state.Version ≠ DB revision
//	state.Materialize ≠ deep clone
//
// # Mesh dependencies
//
// Production code in this package does not import sibling public packages
// (including immutable) and does not import internal/core or
// internal/engine. Mesh sibling law is absolute for production imports:
// state ↛ immutable (and the reverse). Tests may import immutable
// to exercise cross-package scenarios; that is outside production deps.
//
// # Panic prefix
//
// Public panic strings use the "state: …" prefix. Internal or sibling
// package panics must not leak through this API.
package state
