---
type: adr
id: adr:workflow:ast-evidence-for-exercised-surface-gate
context: workflow
title: AST evidence for the exercised-surface gate
status: Accepted
tags: [workflow, archtest, ast, surface, consumers, gates]
deciders: [alvadorn]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:composition-and-resource-safety, adr:workflow:explicit-resilience-as-action-transformation, adr:workflow:sequential-traverse-over-slices, tdd:workflow:core]
date: 2026-10-23
---
# AST evidence for the exercised-surface gate

## Context and problem

`TestSurfaceIsExercised` exists to keep an export from entering the surface by symmetry alone: every symbol needs a consumer outside its own package
(`tdd:workflow:core`). The original gate concatenated the text of all consumer files
and looked for `package.Name` with `strings.Contains`. Raw text does not distinguish use from mention, and the gate accepted evidence that is not consumption:

- **prefix:** `workflow.MapFallible` contains `workflow.Map`, so `Map` looked exercised without an example of its own; the missing `ExampleMap` stayed hidden because of this;
- **comment and string:** `// action.Tap` or `"action.Tap"` counted as use;
- **foreign package:** `other/action.Tap` or an identifier `xaction.Tap` contain the same substring;
- **shadowed identifier:** a local variable named `action` with a field `Tap` counted as the package;
- **false negative:** an aliased import (`act "…/workflow/action"`) uses the symbol and does not match the text `action.Tap`.

The simplicity criterion only holds if the evidence is a real use of the exported symbol of that import path.

## Decision drivers

- evidence identified by exact import path and exact symbol, not by text;
- no new package, no dependency, and no change to the module's public surface
  (`go.mod` still has no `require`);
- the legitimate consumer set must not shrink by accident;
- carrier types returned by a consumed function (`Pair`, `AttemptResult`,
  `RetryPolicy.Backoff`) must not require a separate textual use, but the reachability rule must not become a way to credit any export;
- the gate stays a gate: it needs negative evidence that it fails.

## Decision

The gate reads consumers with `go/parser` (`parser.SkipObjectResolution`) and counts a symbol only when there is a `SelectorExpr` whose qualifier is an identifier bound, by import, to the tracked package, and whose `Sel` is exactly the exported name. The gate tracks the module's five production packages
(`workflow`, `action`, `plan`, `resilience`, `transaction`). Evidence keys are
`importpath + symbol`.

### Qualifier binding

- the local name comes from the import: the explicit alias when there is one; otherwise the known package name of the tracked package;
- a blank import (`_`) is ignored; a dot import (`.`) of a tracked package makes the gate fail explicitly, because an unqualified use is not verifiable by selector;
- a qualifier name only counts outside the spans in which a local identifier shadows it. The gate computes those lexical spans only for names that collide with tracked aliases, following Go's scoping rules: `var`/`const` start at the end of each `ValueSpec`; `:=` starts at the end of the assignment and ends at the end of the block; a local type starts at the `TypeSpec` name; type parameters span the whole signature; receiver, parameter, and result names span the body of `FuncDecl` and `FuncLit`, not the signature; `range :=` spans only the body; the `:=` guard of a `type switch` spans per clause, and the `:=` receive in `select` spans to the end of its clause. A struct field, a label, a composite-literal key, and a parameter name of a standalone `FuncType` bind nothing;
- using the parser's `Obj`/`Scope` was ruled out as the mechanism: `ast.Object` is marked deprecated in the standard library. This decision does not claim the linter rejected it; it chooses the non-deprecated path independently of that.

### Carrier reachability

Direct evidence remains the exact selector. The reachability rule is narrow and local to the signature:

- from an **exported function actually consumed**, follow local named types appearing in its **results**, including nested ones
  (`Action[Pair[A, B]]`, `Action[AttemptResult[T]]`);
- from a **consumed type**, follow the underlying type and the **types** of exported fields, recursively (`RetryPolicy` reaches `Backoff`);
- field names, function parameters, unexported fields and methods, identifiers in bodies, and unconsumed declarations do not count;
- a type-parameter binding is not a package type that happens to share the name;
- a qualified result from another package is not credited.

With this, an unconsumed function stays unconsumed: `Pair` is reached only via `Zip`, `AttemptResult` only via `Attempt`, `Backoff` only via `RetryPolicy`. Constants require a direct selector. The carriers stay public and the surface does not change.

### Consumer set

The gate reads these consumers as AST:

- production `.go` files of this module: a sibling package using the symbol counts, as `TestSurfaceIsExercised` already documents;
- the module's own external `example_test.go` files. They are compiled, executable examples in `package x_test`, that is, real use made through the public API, and are therefore accepted.

`internal/archtest` and the module's ordinary unit tests stay out: a symbol used only by the tests written for it has no consumer, and counting those tests would make the gate vacuous. The filter uses paths relative to the module, and `go vet ./...` also verifies that the consumers compile.

## Alternatives considered

- **Keeping the raw `strings.Contains`.** Rejected: it is the cause of the false positives above and, in practice, already approved an export without an example.
- **Type loading (`go/types` or `golang.org/x/tools/go/packages`).** Would give exact scope, alias, and dot-import resolution. Rejected: it would bring an external dependency or a package loader into a gate, with environment cost and fragility, and would break the decision that the module declares no `require`. The extra information does not change the outcome of a consumption gate.
- **AST without types, with limited lexical scope (chosen).** Preserves exact import path, alias, and shadowing without a loader or dependency, in the style of the parsers `support_test.go` and `context_test.go` already use.
- **Rejecting the module's `example_test.go` files.** Rejected: compiled examples are executable documentation. The gate requires an exact selector and rejects comments and strings as evidence.
- **Requiring a `var _ = package.Name` inside `internal/archtest`.** Rejected: the gate itself would be the consumer, exactly the vacuous pass that excluding `internal/archtest` avoids.

## Consequences

`Map` is no longer credited by `MapFallible`, and a use in a comment, string, foreign package, or shadowed identifier stops counting. Aliased imports now count. The gates gain deterministic negative cases: comment, string, prefix, foreign package, blank import, dot import, alias, generics, shadowing (before, inside the very initializer, after, restored on block exit, grouped `var`, parameters against a valid selector in the signature, nested `FuncLit`,
`if`/`for`/`switch`/`range`/`select`/`type switch`, type parameters, local type,
label, and field name) and the carrier reachability rule.

**Accepted costs.**

- The analysis is syntactic and type-free. The gate proves that a selector exists for the correct import path and symbol, not that the use type-checks or that the code runs: selecting a symbol is not dynamic execution. `go vet ./...` checks types, and `go test ./...` runs the examples with declared output.
- The lexical-scope computation is bespoke code that approximates Go's rules only for names colliding with tracked aliases. There is no general symbol table, loader, or framework; redeclaring an already-local name is handled by interval union without type checking, and a collision between an import name and a package name fails separate compilation, not the gate.
- The reachability rule covers only the local signature (function result, type, and exported fields); it is not a general proof that every type is exercised.
- The consumer set accepts the module's own `example_test.go` files and production sibling packages: the symbol's author can write the consumer. The barrier is requiring a real use, not a comment or text.

No public surface is removed or added: `AcquireUseRelease`, `Attempt`, and `AttemptResult` remain, and `Retry` does not use `errors.Join`. The change touches only `internal/archtest`, the compiled example `ExampleMap`, and documentation.

## Non-goals

Type or dependency loading, binding via dot import of a tracked package, a general symbol table, a new consumer prohibition, a fixed carrier allowlist, proof of dynamic execution, and any change to `action`, `plan`, `resilience`, `transaction`, or the root package.
