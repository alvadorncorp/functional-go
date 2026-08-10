---
type: adr
id: adr:functional:007
context: functional
title: Values family dependency boundary
status: Aceita
tags: [functional, values, generics, dependencies]
relates_to: [adr:functional:001, adr:functional:003, adr:functional:006]
---

# ADR-007: Values family dependency boundary

## Context and problem

The functional module needs five small algebraic value types without coupling
them to persistent data-structure machinery, the state runtime, internal
ownership/engine packages, reflection, unsafe, or domain modeling. A legacy
`functional/optional` package already exists and remains a separate public API.

Validated must accumulate at least one error. Reusing the new NonEmpty type
creates one intentional public-package edge, which is narrower than relaxing
the module's existing sibling-import rule globally.

## Decision drivers

- Preserve useful, documented zero values and value-type cost semantics.
- Keep Optional, Result, Either, and NonEmpty standard-library-only.
- Permit Validated to depend only on NonEmpty.
- Keep existing core, engine, persistent, state, valueobject, and optional APIs untouched.
- Enforce the dependency matrix and no-reflect/no-unsafe rule mechanically.

## Options considered

### Put all types in one `values` package

Rejected: it mixes unrelated public symbols and removes dependency boundaries.

### Move or alias the existing `functional/optional`

Rejected: it changes an established public contract and creates migration work
not authorized by this feature.

### Add five nested packages with one allow-listed edge (chosen)

Chosen: create `functional/values/optional`, `result`, `either`, `nonempty`,
and `validated`. The first four use only the standard library. Validated may
import exactly `functional/values/nonempty`. No root `functional/values` Go
package or re-export facade is created.

## Decision

The values dependency DAG is:

```text
optional   -> stdlib only
result     -> stdlib only
either     -> stdlib only
nonempty   -> stdlib only
validated  -> nonempty
```

All values packages are forbidden from importing legacy public siblings,
`internal/core`, `internal/engine`, `persistent`, `state`, `reflect`, or
`unsafe`.

Architecture tests enumerate all five packages, allow only the directed edge
above, reject `reflect` and `unsafe` in production imports, and keep `go.mod`
free of `require` and `replace` directives.

## Consequences

The feature is additive and rollback is deletion of its new package/docs files
plus reversion of the narrow archtest and documentation inventory changes.
Validated can accumulate `NonEmpty[E]` without a public Semigroup abstraction or
any dependency on runtime layers.

> **Later note (ADR-010):** the parallel top-level `functional/optional` package
> and `functional/valueobject` were removed. `values/optional` is the sole
> public Optional path.
