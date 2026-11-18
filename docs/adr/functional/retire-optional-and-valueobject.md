---
type: adr
id: adr:functional:retire-optional-and-valueobject
context: functional
title: Retire top-level optional and valueobject packages
status: Accepted
tags: [functional, optional, valueobject, dependencies, public-api]
deciders: [alvadorn]
related: [adr:functional:optional-representation-and-zero-value, adr:functional:zero-dependency-boundary, adr:functional:state-semantic-layer, adr:functional:values-family-dependency-boundary]
date: 2026-08-10
---
# Retire top-level optional and valueobject packages

## Context and problem

The `functional` module exposed two public packages with no production
consumers in the monorepo:

- `go.alvadorn.org/functional/optional` — the legacy
  `Optional[T]` API, kept in parallel to `values/optional` (values family
  dependency boundary).
- `go.alvadorn.org/functional/valueobject` — a reflection-based
  test helper (`Check`/`Inspect`), unused in production.

Having two Optionals and a test helper in the public inventory grew the
surface and the documentation with no operational value in the current
graph (`fold` and the rest of the family use `values/*`).

## Decision drivers

- Zero production imports in the monorepo for both paths.
- `values/optional` is already the canonical presence type (the functional
  primitives decision / fold).
- Removal is breaking for external consumers of the legacy path, if any exist.
- Historical ADRs (Optional representation, values family, …) remain as
  records; the live inventory stops listing the packages.

## Options considered

### Option 1: Keep both as legacy

Preserves full compatibility, but keeps duplication and maintenance docs
with no internal consumer.

### Option 2: Deprecate now, remove in a future major

Safer for unknown external consumers; delays the cleanup.

### Option 3: Remove now (chosen)

Delete the packages, the micro-docs, and the live references; record the
decision here. Accept the breaking change in the next release of the
`functional` group.

## Decision

Remove `functional/optional` and `functional/valueobject` and every piece of
product documentation that describes them as current packages. Update the
archtest, the indexes, and the overview. This ADR is the source of record
for the retirement decision.

## Consequences

### Positive

- One public Optional: `values/optional`.
- Smaller public surface and archtest inventory.
- Docs and examples aligned with the real graph (fold → values/optional).

### Negative / trade-offs

- Breaking change: imports of
  `go.alvadorn.org/functional/optional` and
  `.../valueobject` stop compiling.
- External consumers must migrate to `values/optional` or copy the
  value-object helper locally.
- The Optional representation ADR and parts of the values family ADR become
  historical with respect to the top-level `optional` path.

## Verification

- `go list ./...` no longer lists `functional/optional` or
  `functional/valueobject`.
- `go test ./internal/archtest` and `make test-unit` / `make lint` in the
  module.
- Live docs have no broken links to `packages/optional.md` /
  `packages/valueobject.md`.
