# functional documentation

The module root [README](../README.md) is a short **index**. Content lives here.

| Path | Purpose |
|---|---|
| [overview.md](overview.md) | Purpose, the language constraint that shapes the API, the zero-dependency boundary, the sibling-import rule |
| [examples.md](examples.md) | Usage sketches (config precedence, fallible parse, immutable collection field, loop-building) |
| [packages/optional.md](packages/optional.md) | Public API of `github.com/alvadorncorp/gopackages/functional/optional` |
| [packages/frozen.md](packages/frozen.md) | Public API of `github.com/alvadorncorp/gopackages/functional/frozen` |
| [packages/persistent.md](packages/persistent.md) | Public API of `github.com/alvadorncorp/gopackages/functional/persistent`, including measured Append-loop vs. Builder benchmarks |
| [packages/valueobject.md](packages/valueobject.md) | Public API of `github.com/alvadorncorp/gopackages/functional/valueobject` |
| [patterns/value-objects.md](patterns/value-objects.md) | What a value object needs from this module vs. what plain Go already covers |
| [adr/functional/](adr/functional/) | Architecture decisions (`adr:functional:NNN`) |

ADRs use AKB YAML frontmatter (`type`, `id`, `context`, `title`, `status`,
`tags`, …). Layout mirrors monorepo AKB conventions **inside this module**
(`docs/adr/<context>/…`).

All guides are written in **English**.

Current decisions:

- [ADR-001 — Optional representation and zero value](adr/functional/adr-001-optional-representation-and-zero-value.md)
- [ADR-002 — Persistent engine swap contract](adr/functional/adr-002-persistent-engine-swap-contract.md)
- [ADR-003 — Zero-dependency boundary](adr/functional/adr-003-zero-dependency-boundary.md)
