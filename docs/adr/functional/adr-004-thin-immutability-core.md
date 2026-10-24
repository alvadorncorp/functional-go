---
type: adr
id: adr:functional:004
context: functional
title: Thin immutability core — ownership, cow, transient
status: Aceita
tags: [functional, core, ownership, cow, transient, immutable]
deciders: [alvadorn]
related: [adr:functional:001, adr:functional:003, adr:functional:005, adr:functional:014]
date: 2026-08-07
---
# ADR-004: Thin immutability core — ownership, cow, transient

> **Supersession (engine claims only):** Baseline B ships structure-sharing
> engines (HAMT + 32-way vector). Representation decisions, seed policy,
> O(1) Freeze publish, and List.Delete O(n) are recorded in
> [ADR-005](adr-005-persistent-engine-family.md). This ADR remains the
> decision record for the **thin core** (ownership / cow / transient).
> Passages below that describe "full-copy engines" are historical context
> from when the core was extracted.

> **Partial amendment (2026-10-23, [ADR-014](adr-014-v1-surface-hygiene.md)):**
> the core now has **three** packages — `ownership`, `cow`, `transient`. The
> `sharing` package (`Detach` / `Published`) and `cow.NeedsCopy` had no
> production caller and were removed; with O(1) Freeze by `Invalidate`, no
> Detach walk exists. "Needs a copy?" is spelled `!ownership.IsOwnedBy(...)`
> (same semantics and argument order). Status stays **Aceita**; the original
> four-package layout appears below only as historical context, marked
> "(historical)".

## Context and problem

`immutable` already ships List/Map/Set with functional updates and Builders.
A previous plan framed the public API primarily as an "engine swap contract"
(upper-bound log-n complexity, HAMT/RRB readiness). That framing overstated
what the module implements and understated what was actually extracted:

- mutation *permission* (who may edit a node in place),
- a single-threaded *batch window* for Builders,
- publish/share primitives that clear exclusive stamps.

Those concerns must stay separate from structure-specific cloning and from
the engines that own representation. Putting clone or reflection into a
shared core would couple every engine to one representation strategy.

## Decision drivers

- Core answers only **"may I mutate under this owner?"** — never "how do I
  deep-copy this tree?"
- Structures (and future engines) own clone: no generic `Clone[T]`, no
  `reflect` in core.
- Builders need a real mutation window that freezes on `Build` without
  walking nodes.
- Archtests must enforce: zero non-stdlib deps; core DAG is a star into
  `ownership`; no `reflect` in core; public leaves other than `immutable` do
  not import core.
- Public panic strings stay `"immutable: ..."`; transient's own panic is
  an internal detail builders must not leak.

## Options considered

### Option 1: No shared core — each collection invents its own tokens

**Pros:** Zero indirection.
**Cons:** Duplicates owner identity, window lifecycle, and publish rules
across List/Map/Set and any future engine; no mechanical DAG or "no reflect
in core" gate.

### Option 2: Fat core with generic clone / reflection

**Pros:** One place to copy anything.
**Cons:** Forces a single representation strategy; pulls `reflect` into the
hottest shared path; fights archtests and zero-deps posture for the
substrate every engine would share.

### Option 3: Thin core packages — ownership, cow, sharing (historical), transient (chosen)

**Pros:** Clear responsibility split; builders already wire
`transient.Window`; future structure-sharing engines can stamp
owners without rewriting permission logic; clone stays at the edge.
**Cons:** Today's full-copy List/Map/Set barely exercise ownership/cow
yet; those packages exist as the substrate, not as a fully wired
HAMT.

## Decision

**Chosen: Option 3.**

### Layout under `functional/internal/core/` (current: three packages)

| Package | Responsibility |
|---|---|
| `ownership` | Exclusive-mutation token. `Owner` equality is **pointer identity** only. `New()` allocates a non-zero-size value (padding byte) so Go does not collapse zero-sized allocations. `IsOwnedBy(node, owner)` is false when either pointer is nil — `(nil, nil)` is never "owned." A nil stamp means published/shared. |
| `cow` | Mutability decision only: `EnsureEditable` (return node if owned, else call caller-supplied `clone`). **No** structure inspection, **no** deep copy, **no** `Clone[T]`. Nil `clone` when a copy is required panics `"cow: clone required but nil"`. (historical: `NeedsCopy`, the negation of `IsOwnedBy`, was removed by ADR-014.) |
| `transient` | Single-threaded mutation window. `Open()` holds a fresh `*ownership.Owner`; open iff owner is non-nil. `Owner()` and `Freeze()` panic `"transient: window used after Freeze"` on nil, zero-value, or already-frozen windows. `Freeze` only clears the window's owner — it does **not** walk collection nodes. |

(historical: `sharing` — `Detach` cleared a node stamp to nil and `Published` reported a nil stamp — was removed by ADR-014.)

### Dependency DAG (star into ownership)

```text
ownership  (no core deps)
    ↑
cow · transient   (each imports only ownership)
```

No peer edges between cow and transient. Core imports neither
`internal/eq` nor any public package. No production `reflect` under
`internal/core`.

### What core is *not*

- Not a generic persistent-engine library and not a HAMT/RRB
  implementation.
- Not the home of `Equal` (that remains `internal/eq`, which may use
  `reflect` and is outside core).
- Not used by public packages that are required to stay off core (archtested).

### Builder ↔ transient wiring (today)

`ListBuilder` / `MapBuilder` / `SetBuilder` each hold a `*transient.Window`
opened on `Builder()`. Methods gate on a live window; `Build()` freezes the
window and nils the field so later methods panic with the public string
`"immutable: builder used after Build"` rather than leaking the transient
message. Builders still mutate a private full-copy buffer; the window is
lifecycle/scope, not a node-stamping engine yet.

### Relationship to public `immutable` complexity docs

List/Map/Set **still full-copy** on each immutable update over a slice or
map. Documented O(log n) upper bounds and a future structure-sharing
engine remain **aspirational public-API headroom**, not a description of
the thin core and not an implemented engine swap. The core is the substrate
that makes a future engine able to share permission/window/publish rules
without inventing them again. Acceptance of any future engine remains the
external `package immutable_test` suite: public behavior only.

## Consequences

**Positive:** One tested place for mutation tokens and batch windows;
engines can grow without pulling clone/reflect into shared code; Builders
already speak the same Open/Freeze lifecycle a transient-capable engine
would need; archtests lock the DAG and the no-reflect rule.

**Negative:** ownership and cow are lightly used by the current
full-copy engines; maintainers must not mistake their presence for a
shipped HAMT. Complexity tables in package docs can still look "ahead" of
today's constants — that is deliberate API headroom, not core's job.

**Neutral:** ADR-002 (engine-swap-as-design-law) is withdrawn; this ADR
replaces that framing with the shipped core layout. ADR-001 (optional) and
ADR-003 (zero-deps / signature restraint) stay in force unchanged.
