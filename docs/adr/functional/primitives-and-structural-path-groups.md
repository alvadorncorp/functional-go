---
type: adr
id: adr:functional:primitives-and-structural-path-groups
context: functional
title: Group public packages under primitives/ and structural/ path prefixes
status: Accepted
tags: [functional, layout, public-api, dependencies]
deciders: [alvadorn]
related: [adr:functional:values-family-dependency-boundary, adr:functional:functional-primitives-and-fold-boundary, adr:functional:structural-primitives-and-iterator-boundary]
date: 2026-08-10
---
# Group public packages under primitives/ and structural/ path prefixes

## Context and problem

After the module expanded (`values/*`, fold, tuple/lazy/iterator/optics), the
list of public packages at the root grew long. Grouping by **role** improves
discoverability without creating a re-export facade or a `runtime/` folder.

## Decision drivers

- Keep one concept ≈ one Go package (no Go package in the group directories).
- Preserve the DAG and the already-archtested edges (only import paths change).
- Avoid `runtime/` — `immutable` and `state` stay at the module root.
- Accept the breaking path change (module still v0).

## Options considered

### Option 1: Keep everything flat at the root

Zero import churn; discovery only through docs.

### Option 2: `primitives/`, `structural/`, and `runtime/`

Puts collections/state under a prefix too; rejected — `immutable`/`state`
share no family law with the lighter packages, and the `runtime` prefix
suggests infrastructure the module does not intend to unify.

### Option 3: Only `primitives/` and `structural/` (chosen)

- `primitives/{function,predicate,order,fold}`
- `structural/{tuple,lazy,iterator,optics}`
- `values/*`, `immutable`, `state` unchanged in relative position

## Decision

Move the eight packages to the prefixes above. Do not create a Go package in
`primitives` or `structural`. Edges:

- `primitives/fold` → `values/optional`, `values/nonempty`
- `structural/iterator` → `structural/tuple`

## Consequences

**Positive:** an index readable by role; paths express intent.

**Negative:** a breaking rename of eight import paths.

**Verification:** `go list ./...`, archtest, `go test ./...`, lint.
