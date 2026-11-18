---
type: adr
id: adr:functional:structural-primitives-and-iterator-boundary
context: functional
title: Tuples, lazy values, iterators, and optics over native Go contracts
status: Accepted
tags: [functional, generics, iterators, concurrency, optics, dependencies]
deciders: [alvadorn]
related: [adr:functional:zero-dependency-boundary, adr:functional:functional-primitives-and-fold-boundary, adr:functional:v1-surface-hygiene]
date: 2026-08-09
---
# Tuples, lazy values, iterators, and optics over native Go contracts

## Context and problem

The `functional` module needs to remove recurring boilerplate for carrying
heterogeneous values, deferring memoized computations, composing lazy
sequences, and updating nested values immutably. Go 1.26.4 already provides
`iter.Seq`, `iter.Seq2`, `iter.Pull`, `slices.Values`, and `slices.Collect`,
but not the Map/Filter/FlatMap/bounds/Zip/fold/search layer required here.

The zero-dependency boundary ADR requires zero external dependencies and an
exact allow-list between public packages. The functional primitives ADR keeps
function, predicate, and order as leaves and authorizes only the two fold
edges to values. The new packages must preserve that discipline without
creating an Iterator type that competes with the stdlib.

## Decision drivers

- Reuse native Go contracts and range-over-function.
- Keep types small, generic, predictable, and free of reflect/unsafe.
- Avoid coupling to immutable, state, modeling, or internals.
- Make any public edge directed, minimal, and mechanically testable.
- Do not create effect abstractions, academic optics, or stateful iterators.

## Options considered

### Option 1: Skip iterator and rely only on the stdlib

Preserves the smallest surface, but repeats Map, Filter, FlatMap, bounds,
Zip, fold, and searches in consumers that already expose `iter.Seq`.

### Option 2: Create an Iterator type with Next/Value

Centralizes operations, but loses direct interoperability with
range-over-func, `slices`, `maps`, and the iterators already exposed by
immutable.

### Option 3: Small packages with iterator as adapters over iter.Seq

Keeps tuple, lazy, and optics as stdlib leaves; iterator receives and returns
`iter.Seq` and imports only tuple, to represent Zip and Enumerate as a single
composable value.

## Decision

**Chosen: Option 3.** Create `tuple`, `lazy`, `iterator`, and `optics` in the
existing module. Do not create a facade at the root package. `tuple`, `lazy`,
and `optics` depend only on the stdlib. `iterator` depends on the stdlib and
exactly on `tuple`.

`iterator` does not define an Iterator type. FromSlice and Collect delegate to
`slices.Values` and `slices.Collect`; the remaining operations preserve
laziness and early termination. Zip uses `iter.Pull`, ends both pulls, and
terminates on the shorter input. The pull order is left then right; if the
right side ends first, at most one unpaired left value may have been consumed.

`lazy` is a handle behind a pointer: it runs at most once, is safe for
concurrent use, memoizes both success and panic like `sync.OnceValue`,
releases the stored initializer function, and replays the same panic payload.
The original implementation used explicit state over `sync.Once`; the
amendment below replaces it with `sync.OnceValue`. `optics` starts only with
a typed Lens and explicit composition.

## Consequences

**Positive:** native interoperability, no intermediate collections, a single
new public edge, and small APIs with executable ceilings.

**Negative:** iterator keeps nominal wrappers for two slices adapters (now
`Deprecated`, see the amendment); Zip inherits the fixed overhead of Pull and
the documented asymmetric consumption; Lazy retains the memoized value or
panic for its own lifetime.

**Neutral / Rollback:** each package is additive. Reverting the phases removes
the corresponding code, guide, and inventory. Removing iterator also removes
the only new edge. Prism, Traversal, async iterators, effects, and caches stay
out.

## Partial amendment (2026-10-23, v1 surface hygiene)

A partial amendment that does not replace this decision; the status remains
**Accepted**. Recorded alongside the [v1 surface hygiene](v1-surface-hygiene.md) decision.

- **`Deprecated` synonyms, no removal.** `iterator.FromSlice` and
  `iterator.FromValues` (replaced by `slices.Values`), `iterator.Collect`
  (replaced by `slices.Collect`), and `tuple.FromValues2` (replaced by
  `Of2`) stay exported with the same semantics; they now carry the
  `Deprecated:` marker. The source adapters remain live views of the
  caller's slice. The rest of the `iterator` and `tuple` API is unchanged.
- **`Lazy` over `sync.OnceValue`.** The original text described explicit
  state protected by `sync.Once` that "mirrors" `sync.OnceValue`. `Lazy` now
  uses `sync.OnceValue` directly. Single execution under concurrency,
  common-panic memoization, and initializer release are preserved. The fix:
  with the manual implementation, a `runtime.Goexit` in the initializer left
  the `Lazy` marked complete with the zero value, and every later `Get`
  silently returned it; with `sync.OnceValue`, that `Get` and all later ones
  panic. Under Go's default `panic(nil)` conversion (main module with
  `go >= 1.21` and `GODEBUG panicnil` unset) the payload is
  `*runtime.PanicNilError`. Under the legacy `panicnil=1` override,
  `sync.OnceValue` follows the runtime: it never returns the zero value as
  initialized, but the payload type is not guaranteed. No change to the
  default handling of `panic(nil)` is claimed.
