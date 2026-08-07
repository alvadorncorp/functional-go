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

---

## What core is for

Core answers only: **may this value be mutated in place under the current
owner, or must the caller copy first?** It does not deep-copy nodes, does
not walk trees on freeze, and does not implement List/Map/Set.

| Package | Role |
|---|---|
| `ownership` | Exclusive-mutation token (`*Owner`); pointer identity only |
| `cow` | "Needs a copy?" / `EnsureEditable` with caller-supplied clone |
| `sharing` | Detach / Published — clear exclusive stamps for sharing |
| `transient` | Single-threaded Open → Owner → Freeze mutation window |

Structure-specific clone logic always lives in the collection or engine
layer. Core never provides `Clone[T]` and never uses `reflect`.

---

## `ownership`

- `Owner` is a pure identity token. Two owners are the same only when they
  are the same pointer.
- `New()` returns a unique `*Owner`. The type embeds an unexported byte so
  the allocation is non-zero-size (Go would otherwise collapse zero-sized
  allocations to one address).
- Nodes in a future structure-sharing engine stamp `*Owner`. **Nil** means
  published/shared — not exclusively owned.
- `IsOwnedBy(nodeOwner, owner)` is true only when both pointers are non-nil
  and equal. `(nil, nil)` is **not** owned.

---

## `cow`

- `NeedsCopy(nodeOwner, owner)` ≡ `!ownership.IsOwnedBy(...)`.
- `EnsureEditable[N](nodeOwner, owner, node, clone)` returns `node` when
  already owned; otherwise calls `clone()`. If a copy is required and
  `clone` is nil, panics `"cow: clone required but nil"`.
- The concrete type of `N` is opaque to core. The caller must stamp the
  clone with `owner` if exclusive mutation is intended afterward.

---

## `sharing`

- Dual of `cow`: leave exclusive mutation rather than enter it.
- `Detach(nodeOwner **Owner)` sets `*nodeOwner` to nil (no-op if already
  nil or the double-pointer is nil).
- `Published(nodeOwner *Owner)` is true when the stamp is nil.
- Does not clone; after detach, any mutation under any owner requires a
  structure-defined copy.

---

## `transient`

- `Window` is a single-threaded mutation scope (**not** concurrency-safe).
- Open window ⇔ non-nil internal owner (`Open()` allocates via
  `ownership.New()`).
- `Owner()` returns the live token; panics
  `"transient: window used after Freeze"` if the window is nil, zero-value,
  or already frozen.
- `Freeze()` permanently closes the window by clearing its owner. It does
  **not** walk collection nodes; stamping/detaching nodes at publish time
  is the collection layer's job.
- A second `Freeze` panics with the same message.

### How `persistent` Builders use it today

Builders open a window on `Builder()`, gate every method on a live window,
and on `Build()` call `Freeze` then nil their `win` field. Public panics
remain `"persistent: builder used after Build"` — the transient panic
string is not exposed to builder callers. The buffered data path is still a
private full-copy slice/map; the window is lifecycle, not a node stamp
graph yet.

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
- `optional`, `frozen`, and `valueobject` must not import core.
- `persistent` may import core (Builders already use `transient`).

Enforced by `functional/internal/archtest`
(`TestCoreDependencyDAG`, `TestCoreNoReflect`,
`TestPublicPackagesDoNotImportCore`, plus the module-wide zero-deps tests).

---

## Related internal package (outside core)

`functional/internal/eq` resolves `Equal` in three tiers (method, `==`,
`reflect.DeepEqual`). It is shared by public packages and **may** use
`reflect`. It is deliberately **not** part of core so the mutation
substrate stays free of reflection.

---

## Related documents

- [Overview](../overview.md)
- [Package API: persistent](persistent.md)
- [ADR-004 — Thin immutability core](../adr/functional/adr-004-thin-immutability-core.md)
- [ADR-003 — Zero-dependency boundary](../adr/functional/adr-003-zero-dependency-boundary.md)
