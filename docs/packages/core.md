# Internal packages: `internal/core`

**Import paths (not public API):**

| Package | Path |
|---|---|
| ownership | `go.alvadorn.org/functional/internal/core/ownership` |
| cow | `go.alvadorn.org/functional/internal/core/cow` |
| transient | `go.alvadorn.org/functional/internal/core/transient` |

Maintainer guide for the thin immutability substrate: three packages
(`ownership`, `cow`, `transient`). Callers outside
`functional` cannot import these packages (`internal/`). For the decision
record see [thin immutability core](../adr/functional/thin-immutability-core.md)
and its partial amendment in
[v1 surface hygiene](../adr/functional/v1-surface-hygiene.md).
Engines (HAMT, vector, wbtree) are separate — see
[persistent engine family](../adr/functional/persistent-engine-family.md),
[sorted map weight balanced engine](../adr/functional/sorted-map-weight-balanced-engine.md), and
[immutable.md](immutable.md).

---

## What core is for

Core answers only: **may this value be mutated in place under the current
owner, or must the caller copy first?** It does not deep-copy nodes, does
not walk trees on freeze, and does not implement List/Map/Set/SortedMap.

| Package | Role |
|---|---|
| `ownership` | Exclusive-mutation token (`*Owner`); pointer identity + Active state |
| `cow` | `EnsureEditable` with caller-supplied clone |
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
  owner fail `IsOwnedBy`, so engines need not walk the tree to clear stamps.
- `IsOwnedBy(nodeOwner, owner)` is true only when both pointers are non-nil,
  equal, **and** `owner` is Active. `(nil, nil)` is **not** owned; an
  invalidated owner is never an editable owner.

---

## `cow`

- `EnsureEditable[N](nodeOwner, owner, node, clone)` returns `node` when
  already owned (`ownership.IsOwnedBy`); otherwise calls `clone()`. If a copy
  is required and `clone` is nil, panics `"cow: clone required but nil"`.
- The concrete type of `N` is opaque to core. The caller must stamp the
  clone with `owner` if exclusive mutation is intended afterward.
- To ask "needs a copy?" directly, engines use `!ownership.IsOwnedBy(...)`.
  After `Invalidate`, `IsOwnedBy` is false even if the pointer stamp still
  matches — engines copy-on-write into the new window.
- There is no `NeedsCopy` function: it was removed as an unused alias of
  `!IsOwnedBy` (`adr:functional:v1-surface-hygiene`).

---

## Removed: `sharing`

The former `sharing` package (`Detach` / `Published`) had no production
caller and was removed (`adr:functional:v1-surface-hygiene`). Freeze is O(1) by `Invalidate`, so inactive
stamps already fail `IsOwnedBy`; no walk clears stamps, and a published node
is simply one whose stamp is nil or inactive.

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
cow · transient
```

- Star into `ownership` only — no peer edge between cow and transient.
- No `reflect` in production files under `internal/core`.
- Core must not import `internal/eq` or any public subpackage.
- Public packages other than `immutable` must not import core.
- `immutable` may import core (Builders use `transient`).
- Engines may import `ownership` / `cow` but **not** `transient`
  (callers pass an owner from a window).

Enforced by `functional/internal/archtest`
(`TestCoreDependencyDAG`, `TestCoreNoReflect`,
`TestPublicPackagesDoNotImportCore`, plus the module-wide zero-deps tests).

---

## Related internal package (outside core)

`functional/internal/eq` resolves `Equal` without `reflect` (method, then
`==`; non-comparable types without `Equal` return false). It is shared by
public packages and is deliberately **not** part of core so the mutation
substrate stays a pure ownership/COW layer.

Engines live under `functional/internal/engine/` (`hamt`, `vector`,
`wbtree`, `keyhash`) — representation, not laws. See
[immutable.md](immutable.md),
[persistent engine family](../adr/functional/persistent-engine-family.md), and
[sorted map weight balanced engine](../adr/functional/sorted-map-weight-balanced-engine.md).

---

## Related documents

- [Overview](../overview.md)
- [Package API: immutable](immutable.md)
- [Thin immutability core](../adr/functional/thin-immutability-core.md)
- [Persistent engine family](../adr/functional/persistent-engine-family.md)
- [v1 surface hygiene (removes `sharing` and `cow.NeedsCopy`)](../adr/functional/v1-surface-hygiene.md)
- [Zero-dependency boundary](../adr/functional/zero-dependency-boundary.md)
