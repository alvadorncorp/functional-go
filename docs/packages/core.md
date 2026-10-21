# Internal packages: `internal/core`

**Import paths (not public API):**

| Package | Path |
|---|---|
| ownership | `github.com/alvadorncorp/gopackages/functional/internal/core/ownership` |
| cow | `github.com/alvadorncorp/gopackages/functional/internal/core/cow` |
| sharing | `github.com/alvadorncorp/gopackages/functional/internal/core/sharing` |
| transient | `github.com/alvadorncorp/gopackages/functional/internal/core/transient` |

Maintainer guide for the thin immutability substrate. Callers outside
`functional` cannot import these packages (`internal/`). For the decision
record see [ADR-004](../adr/functional/adr-004-thin-immutability-core.md).
Engines (HAMT, vector, wbtree) are separate — see
[ADR-005](../adr/functional/adr-005-persistent-engine-family.md),
[ADR-013](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md), and
[immutable.md](immutable.md).

---

## What core is for

Core answers only: **may this value be mutated in place under the current
owner, or must the caller copy first?** It does not deep-copy nodes, does
not walk trees on freeze, and does not implement List/Map/Set/SortedMap.

| Package | Role |
|---|---|
| `ownership` | Exclusive-mutation token (`*Owner`); pointer identity + Active state |
| `cow` | "Needs a copy?" / `EnsureEditable` with caller-supplied clone |
| `sharing` | Detach / Published — clear exclusive stamps |
| `transient` | Single-threaded Open → Owner → Freeze mutation window |

**Laws vs representation:** core = mutation permission laws; engines =
structure representation (path-copy, branching factor, hash seed). Structure-
specific clone logic always lives in the collection or engine layer. Core
never provides `Clone[T]` and never uses `reflect`.

---

## `ownership`

- `Owner` is a pure identity token with an **Active** flag. Two owners are
  the same only when they are the same pointer **and** the token is still
  Active.
- `New()` returns a unique, active `*Owner`.
- Nodes stamp `*Owner`. **Nil** means published/shared — not exclusively
  owned (born published).
- `Active()` is true for a non-nil token that has not been invalidated.
- `Invalidate()` permanently deactivates the token (idempotent; nil-safe).
  **Freeze uses this for O(1) publish:** residual stamps of an inactive
  owner fail `IsOwnedBy`, so engines need not walk the tree to Detach.
- `IsOwnedBy(nodeOwner, owner)` is true only when both pointers are non-nil,
  equal, **and** `owner` is Active. `(nil, nil)` is **not** owned; an
  invalidated owner is never an editable owner.

---

## `cow`

- `NeedsCopy(nodeOwner, owner)` ≡ `!ownership.IsOwnedBy(...)`.
- `EnsureEditable[N](nodeOwner, owner, node, clone)` returns `node` when
  already owned; otherwise calls `clone()`. If a copy is required and
  `clone` is nil, panics `"cow: clone required but nil"`.
- The concrete type of `N` is opaque to core. The caller must stamp the
  clone with `owner` if exclusive mutation is intended afterward.
- After `Invalidate`, `NeedsCopy` is true even if the pointer stamp still
  matches — engines copy-on-write into the new window.

---

## `sharing`

- Dual of `cow`: leave exclusive mutation rather than enter it.
- `Detach(nodeOwner **Owner)` sets `*nodeOwner` to nil (no-op if already
  nil or the double-pointer is nil).
- `Published(nodeOwner *Owner)` is true when the stamp is nil.
- Does not clone; after detach, any mutation under any owner requires a
  structure-defined copy.
- With O(1) Freeze via `Invalidate`, Detach walks are optional: inactive
  stamps already fail `IsOwnedBy`.

---

## `transient`

- `Window` is a single-threaded mutation scope (**not** concurrency-safe).
- Open window ⇔ non-nil internal owner (`Open()` allocates via
  `ownership.New()`).
- `Owner()` returns the live token; panics
  `"transient: window used after Freeze"` if the window is nil, zero-value,
  or already frozen.
- `Freeze()` permanently closes the window in **O(1)**: it
  `Invalidate()`s the owner token and clears the window's owner pointer.
  It does **not** walk collection nodes; residual inactive stamps are
  treated as published by the next window.
- A second `Freeze` panics with the same message.

### How `immutable` Builders use it

Builders open a window on `Builder()`, gate every method on a live window,
and pass `win.Owner()` into engine transient methods (`TransientSet`,
`TransientAppend`, …). On `Build()` they call `Freeze` then nil their `win`
field. Public panics remain `"immutable: builder used after Build"` — the
transient panic string is not exposed to builder callers. The data path is
the structure-sharing engine under exclusive ownership, not a private
full-copy buffer.

---

## Dependency rules (archtested)

```text
ownership
    ↑
cow · transient · sharing
```

- Star into `ownership` only — no peer edges among cow/transient/sharing.
- No `reflect` in production files under `internal/core`.
- Core must not import `internal/eq` or any public subpackage.
- Public packages other than `immutable` must not import core.
- `immutable` may import core (Builders use `transient`).
- Engines may import `ownership` / `cow` / `sharing` but **not** `transient`
  (callers pass an owner from a window).

Enforced by `functional/internal/archtest`
(`TestCoreDependencyDAG`, `TestCoreNoReflect`,
`TestPublicPackagesDoNotImportCore`, plus the module-wide zero-deps tests).

---

## Related internal package (outside core)

`functional/internal/eq` resolves `Equal` without `reflect` (method, then
`==`; non-comparable types without `Equal` return false). It is shared by
public packages and is deliberately **not** part of core so the mutation
substrate stays a pure ownership/COW/sharing layer.

Engines live under `functional/internal/engine/` (`hamt`, `vector`,
`wbtree`, `keyhash`) — representation, not laws. See
[immutable.md](immutable.md),
[ADR-005](../adr/functional/adr-005-persistent-engine-family.md), and
[ADR-013](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md).

---

## Related documents

- [Overview](../overview.md)
- [Package API: immutable](immutable.md)
- [ADR-004 — Thin immutability core](../adr/functional/adr-004-thin-immutability-core.md)
- [ADR-005 — Persistent engine family](../adr/functional/adr-005-persistent-engine-family.md)
- [ADR-003 — Zero-dependency boundary](../adr/functional/adr-003-zero-dependency-boundary.md)
