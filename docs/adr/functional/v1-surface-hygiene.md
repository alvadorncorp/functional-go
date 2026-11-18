---
type: adr
id: adr:functional:v1-surface-hygiene
context: functional
title: v1 surface hygiene — Deprecated synonyms, lean core, native Lazy, frozen state
status: Accepted
tags: [functional, v1, deprecation, lazy, hamt, core, state, iterator, predicate, tuple]
deciders: [alvadorn]
related: [adr:functional:thin-immutability-core, adr:functional:state-semantic-layer, adr:functional:structural-primitives-and-iterator-boundary]
date: 2026-10-23
---
# v1 surface hygiene — Deprecated synonyms, lean core, native Lazy, frozen state

## Context and problem

After the collections phases, `functional` had accumulated residue that does
not change the public contract but confuses whoever reads or maintains the
module:

- `structural/iterator.FromSlice`, `FromValues`, and `Collect`,
  `primitives/predicate.All`/`Any`/`None`, and `structural/tuple.FromValues2`
  are thin synonyms of stdlib functions or of another constructor in their own
  package (`slices.Values`, `slices.Collect`, `slices.ContainsFunc`, `Of2`).
- `internal/core/sharing` (`Detach`/`Published`) and `cow.NeedsCopy` have no
  production caller: the `O(1)` Freeze via `Invalidate` (persistent engine
  family) made the `Detach` walk unnecessary, and `NeedsCopy` is just
  `!ownership.IsOwnedBy`.
- The internal HAMT keeps `merge` paths that no caller reaches (equal hashes,
  `shift` beyond the maximum) and, after `Delete`, leaves children with a
  single terminal element inside intermediate nodes.
- `structural/lazy` reimplements by hand what `sync.OnceValue` already does,
  with a real divergence: a `runtime.Goexit` inside the initializer leaves the
  `Lazy` marked complete with the zero value, and every later `Get` returns
  that zero with no signal.
- `state` is shipped and stable, but no document says whether the package
  evolves, gets replaced, or gets removed.

The owner approved, on 2026-10-23, only these items. The change must
preserve the whole v1 public API (every exported name keeps compiling) and
the dependency boundary of the zero-dependency boundary ADR.

## Decision drivers

- No exported v1 name is removed; removal would break consumers with no
  proportional gain.
- Prefer native stdlib contracts over manual reimplementations (tuples,
  lazy, iterators, and optics ADR).
- Signal consumers toward the preferred path without suppressing lint or
  relaxing gates.
- Remove internal code with no caller and unreachable branches, migrating
  every caller completely, with no shims.
- HAMT compression only where it is safe: the bitmap depends on the position
  (`shift`) and cannot move up a level.
- Do not claim unmeasured behavior or cost.

## Options considered

### Option 1: Keep everything as is

No behavioral risk. The synonyms stay announced as the preferred API, dead
code keeps demanding reading and tests, `Lazy` keeps the silent zero after
`Goexit`, and the future of `state` stays ambiguous.

### Option 2: Remove synonyms and `state` at the v1 line

Gives the minimal surface. Breaks any existing consumer of the removed names
and contradicts the v1 compatibility promise; the owner did not approve
removals.

### Option 3: Mark synonyms as `Deprecated`, clean internals, adopt `sync.OnceValue`, and freeze `state` (chosen)

Keeps every exported name, guides consumers to the stdlib through the
standard `Deprecated:` marker, removes only internal code with no caller,
and swaps `Lazy`'s manual machinery for the native primitive.

## Decision

**Chosen: Option 3**, because it is the only one that satisfies
"no v1 name removed" and "one preferred path per capability" at the same
time.

1. **`Deprecated` synonyms (no removal).** `iterator.FromSlice` and
   `iterator.FromValues` point to `slices.Values`; `iterator.Collect` points
   to `slices.Collect`; `predicate.All`/`Any`/`None` point to
   `slices.ContainsFunc` (and to the `iterator.All`/`Any` quantifiers, for
   whoever already imports `iterator`); `tuple.FromValues2` points to `Of2`.
   The semantics do not change: the source adapters remain **live views** of
   the caller's slice, with no copy. `predicate` stays a stdlib-only leaf and
   does not import `iterator`. The examples, benchmarks, and real consumers
   migrate to the replacements.
2. **`state` frozen, not `Deprecated`.** The existing API is supported and
   stable, with no new surface, no `Deprecated` marker on the package, and no
   promise of removal or of a v2 (partial amendment in the state semantic
   layer ADR, which stays `Accepted`).
3. **Lean core.** The `internal/core/sharing` package and `cow.NeedsCopy`
   are removed; the core goes from four to three packages (`ownership`,
   `cow`, `transient`). Whoever needed `NeedsCopy` uses
   `!ownership.IsOwnedBy`, with the same semantics and the same argument
   order. The stamps of inactive owners stay as they are (`O(1)` Freeze via
   invalidation). This is a partial amendment of the thin immutability core
   ADR, which stays `Accepted`.
4. **HAMT without dead branches, with terminal compression.** `merge` is
   only called with distinct hashes; the equal-hash and beyond-maximum-`shift`
   branches become `panic("hamt: unreachable merge")`, and `joinCollision`
   and `collectEntries` go away. In `Delete` (pure and transient), when a
   child ends up holding only a leaf or a collision, the parent stores that
   terminal directly. **Never** promote a bitmap, because it depends on the
   `shift`. Count, owner, and earlier snapshots are preserved. None of this
   is visible in the public API.
5. **`Lazy` over `sync.OnceValue`.** `Lazy[T]` now stores a getter
   `func() T` created by `sync.OnceValue`. `New` with a nil initializer keeps
   the same panic; the zero-value `Lazy` and a nil `*Lazy` stay documented as
   invalid, with no extra guard. Preserved: single execution even with
   concurrent callers, caching of the value and of the common panic payload,
   and the release of the initializer function. The only approved behavioral
   difference is `runtime.Goexit` inside the initializer: the `Lazy` stops
   recording the zero value as success — that `Get` and every later one
   panic. Under Go's default `panic(nil)` conversion (main module with
   `go >= 1.21` and `GODEBUG panicnil` unset) the payload is
   `*runtime.PanicNilError`; under the legacy `panicnil=1` override,
   `sync.OnceValue` follows the runtime and the payload type is not
   guaranteed, although the zero value is never returned as success. There is
   no change to the default handling of `panic(nil)`: modern Go already
   converts it into `*runtime.PanicNilError` before `recover`, and that
   behavior is not the object of this decision. This point corrects the
   description in the tuples/iterators ADR (partial amendment, which stays
   `Accepted`).

Rejected alternatives per item: removing the synonyms (Option 2); promoting
bitmap nodes with a single child to the parent in the HAMT (incorrect, the
bitmap depends on the `shift`); an explicit guard for a zero or nil `Lazy`
and `Lazy` options (scope not approved); marking `state` as `Deprecated` (the
package is not being retired).

## Consequences

**Positive:**

- One preferred path per capability, announced by the `Deprecated:` marker,
  with no v1 name removed.
- The core ends with three packages and no caller-less API; the internal
  dependency graph shrinks (`hamt` stops importing `sharing`).
- `Lazy` loses the manual `Once`/value/panic machinery and closes the silent
  zero after `Goexit`.
- The HAMT structure after `Delete` becomes canonical for terminals, with no
  extra transients.

**Negative:**

- The `Deprecated:` markers make the linter (SA1019) point at existing uses
  in consumers, which must migrate or accept the warning. The names stay
  available, so the synonyms remain surface to maintain.
- Compatibility tests for the `Deprecated` names must live in the package
  that defines them, to avoid violating the lint, and must not become
  tautological tests of the stdlib.
- A `Goexit` inside a `Lazy` initializer now has a different observable
  effect than before (a panic on that `Get` and the later ones instead of the
  silent zero). This is an approved behavior change, not only internal.
- Observed measurement of `Lazy` with `sync.OnceValue`, restricted to the
  existing benchmarks (`go -C functional test -run '^$' -bench
  '^BenchmarkLazy(ColdGet|WarmGet)$' -benchmem ./structural/lazy`, Go 1.27.3,
  darwin/arm64, Apple M2 Max, **one sample** each): `BenchmarkLazyColdGet`
  went from 64 B/op and 1 alloc/op to 48 B/op and 1 alloc/op (103.7 ns/op
  after), and `BenchmarkLazyWarmGet` stayed at 0 B/op and 0 allocs/op in both
  cases (7.087 ns/op after). This is an allocation comparison on those two
  benchmarks, not a statistical analysis of latency and not a general
  allocation guarantee of the API.
- The HAMT now relies on the "merge only receives distinct hashes"
  invariant from its callers; a violation of that invariant becomes a panic
  instead of a silent branch.
- A frozen, non-`Deprecated` `state` gives the consumer no signal about the
  package's long-term direction.

**Neutral / Follow-up:**

- The thin immutability core and tuples/iterators ADRs receive partial
  amendments and stay `Accepted`; the state semantic layer ADR receives the
  frozen-`state` amendment. No ADR is superseded as a whole.
- The guides for `iterator`, `predicate`, `tuple`, `lazy`, `fold`, and
  `core`, the overview, and the indexes now describe the post-decision state.
- Out of scope: any removal of a v1 name, new APIs, `Lazy` options, and
  changes to `modeling` or `workflow`.

---

*Related ADRs: adr:functional:thin-immutability-core, adr:functional:state-semantic-layer, adr:functional:structural-primitives-and-iterator-boundary*
