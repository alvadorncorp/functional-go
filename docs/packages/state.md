# Package `state`

**Import path:** `github.com/alvadorncorp/gopackages/functional/state`

Semantic evolution layer for values of type `T` (any). `State[T]` holds a
base value and an immutable **plan** of pure mutations; each evolution
returns a new `State` rather than mutating the receiver. The package owns
lineage, version identity, single-use transactions, and structural diffs for
that pure `T`. It does **not** own domain rules, validation, or persistence
engines.

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Construction | `Of[T]`, zero `State[T]` |
| Pure evolution | `Mutation[T]`, `State.Apply` |
| Materialization | `State.Materialize`, `State.Snapshot` / `Snapshot[T]` |
| Lineage identity | unexported lineage token; `Version` |
| Single-use batching | `State.Begin`, `Transaction[T]` (`Apply`, `Commit`) |
| Structural plan diff | `DiffBetween`, `Diff[T]` (`From`, `To`, `Len`, `Apply`) |

---

## Distinctions (read these first)

| Concept | Is | Is not |
|---|---|---|
| `state.Transaction` | Single-use plan accumulator over a captured `State` | `persistent.Builder` (exclusive-owner transient over a collection engine) |
| `state.Diff` | Structural plan steps between two States of one lineage | A domain event, audit log entry, or CQRS message |
| `state.Version` | Monotonic lineage counter on a pure plan (`uint64`) | A database revision, row version, or optimistic-lock token |
| `state.Materialize` | Re-apply the full plan oldest→newest (no cache) | A deep clone of `T`; immutability is only as deep as `T` |

```
state.Transaction ≠ persistent.Builder
state.Diff ≠ Domain Event
state.Version ≠ DB revision
state.Materialize ≠ deep clone
```

---

## The contract

1. **`T` is opaque.** `state` never interprets fields of `T` and never
   imports domain packages. Prefer immutable or value-object `T` when deep
   immutability of the payload is required.
2. **`Apply` is lazy.** It allocates one reverse-linked plan node and does
   **not** invoke the mutation. `Materialize` (and `Snapshot`) run the plan.
3. **No materialize cache.** Every `Materialize` re-applies the full plan.
   Callers that need a frozen value once should use `Snapshot`.
4. **`Apply` is O(1)** plan growth (one node prepend). `Materialize` is
   linear in plan length (walk + temporary steps slice + mutation work).
5. **Lineage is pointer identity.** Each `Of` allocates a fresh, non-zero-size
   lineage token; derived States share it. Zero States use one canonical
   zero-value lineage, so `DiffBetween(zero, zero.Apply(...))` is valid.
   `DiffBetween` / `Diff.Apply` panic with `"state: unrelated lineages"`
   across `Of` roots or non-descendant plans.
6. **Version advances by one per `Apply`** on the returned State (or on the
   transaction working copy). Sibling branches from the same parent both
   advance independently; Version does not encode plan shape.
7. **Transaction is single-use and not concurrency-safe.** After `Commit`,
   further `Apply` or `Commit` panics `"state: transaction used after Commit"`.
   `Commit` does not call `Materialize`. An empty Commit returns a State
   equivalent to the Begin snapshot (same base, plan, Version).
8. **Mesh dependencies.** Production `state` does not import sibling public
   packages (`optional`, `persistent`, `valueobject`) and does not import
   `internal/core` or `internal/engine`. Tests may import `persistent` to
   exercise composition; that is outside production deps.
9. **Public panics** use the `"state: …"` prefix only.
10. **No production `reflect` / `unsafe`.** Diff is structural over the plan,
    not a reflective field walk of `T`.

---

## API

### Construction and evolution

| Function / method | Behaviour |
|---|---|
| `Of[T](value T) State[T]` | Base value, empty plan, Version 0, fresh lineage. |
| zero `State[T]` | Valid canonical empty root: `Materialize` returns zero `T`, Version 0; `Apply` can be diffed from that root. |
| `(s State[T]) Version() Version` | Monotonic counter; `Of` / zero report 0. |
| `(s State[T]) Apply(m Mutation[T]) State[T]` | New State with `m` prepended; receiver unchanged; `m` not invoked; Version +1. |
| `(s State[T]) Materialize() T` | Apply plan oldest→newest to base; full re-apply every call. |
| `(s State[T]) Snapshot() Snapshot[T]` | One-shot materialization: frozen `Version` + `Value`. |

`Mutation[T]` is `func(T) T`. Callers must treat it as free of side effects
on shared state: `Materialize` may invoke it any number of times.

### Transaction

| Function / method | Behaviour |
|---|---|
| `(s State[T]) Begin() *Transaction[T]` | Capture base, plan head, Version. Parent unchanged. |
| `(tx *Transaction[T]) Apply(m Mutation[T]) *Transaction[T]` | Record `m` on the working plan (chainable). |
| `(tx *Transaction[T]) Commit() State[T]` | Return working State; freeze tx. No Materialize. |

### Diff

| Function / method | Behaviour |
|---|---|
| `DiffBetween[T](from, to State[T]) Diff[T]` | Plan steps from `from` to `to` (same lineage; `to` descendant of `from`). |
| `(d Diff[T]) From() / To() Version` | Endpoint versions captured at DiffBetween. |
| `(d Diff[T]) Len() int` | Number of mutations in the Diff. |
| `(d Diff[T]) Apply(base State[T]) State[T]` | Replay steps onto `base` (must match From / head / lineage). Lazy. |

---

## Semantics notes

### Lazy Apply, reverse-linked plan

`Apply` prepends a plan node (`head` → newer; `older` walks toward base).
Sibling branches share the parent's plan prefix by pointer. That sharing is
why `Apply` stays O(1) and why `DiffBetween` can walk `to`'s chain until it
finds `from`'s head.

### Snapshot vs Materialize

`Materialize` is pure re-computation. `Snapshot` is the "freeze once" helper:
it stores the Version and the materialized Value at call time. Evolving the
source State afterward does not change an already-taken Snapshot.

### Composing with `persistent` (application layer)

Production `state` never imports `persistent`. Application code may still
use `State[MyType]` where `MyType` fields are `persistent.Map` / `Set` /
`List`, evolving those fields inside pure `Mutation`s. Integration tests in
this package demonstrate that composition; it is not a mesh exception.

---

## Measured: Apply vs Materialize

```sh
cd functional
go test ./state/ -run '^$' -bench . -benchmem
```

- **`BenchmarkStateApply`:** marginal cost of one `Apply` on a plan of size
  `n`. With reverse-linked prepend, **allocs/op stays ~constant in n**.
- **`BenchmarkStateMaterialize`:** full re-apply of a plan of size `n`
  (1 / 10 / 100 / 1K). Cost scales with plan length.
- **`BenchmarkTransactionApplyCommit`:** batch of Applies via Tx vs sequential
  note-style `State.Apply`.
- **`BenchmarkPersistentInState`:** optional cheap probe with a `Map` field
  (test-only import of `persistent`).

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [Package API: persistent](persistent.md) (collections; not imported by state)
- [ADR-006 — State semantic layer](../adr/functional/adr-006-state-semantic-layer.md)
- [ADR-003 — Zero-dependency boundary](../adr/functional/adr-003-zero-dependency-boundary.md)
