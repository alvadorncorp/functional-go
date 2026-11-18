---
type: tdd
id: tdd:workflow:core
context: workflow
title: Workflow module core
status: Accepted
tags: [workflow, core, module, architecture, archtest]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-integration-without-effect-duplication, adr:workflow:composition-and-resource-safety, adr:workflow:plan-as-explicit-non-executable-workflow-description, adr:workflow:explicit-resilience-as-action-transformation, adr:workflow:ast-evidence-for-exercised-surface-gate, tdd:workflow:action, tdd:workflow:composition, tdd:workflow:plan-resilience]
date: 2026-08-11
---
# TDD: Workflow module core

## Technical scope

The module keeps five production packages and one internal gate package:
`workflow`, `workflow/action`, `workflow/plan`, `workflow/resilience`,
`workflow/transaction`, and `workflow/internal/archtest`.

```text
workflow ──► action ◄── transaction
                ▲
                └────── resilience

plan   (leaf, no intra-module dependency)
```

`action` is the sink and imports exactly `context` and `errors`. The second
dependency exists only to preserve two independent failures in `Ensure` and
`Bracket`; both stay in the standard library. `plan` imports only `slices` and
`strings`; `resilience` imports exactly `context`, `action`, and `time`. The
`plan` and `resilience` contracts are in
[`plan-resilience.md`](plan-resilience.md). `go.mod` contains no `require` or
`replace`.

## Public contracts

```go
type Workflow[I, O any] func(I) action.Action[O]
type Action[T any] func(context.Context) (T, error)
type Boundary[S any] func(context.Context, func(context.Context, S) error) error
```

`TestExactPublicSurface` pins the complete surfaces; a change to a name,
signature, field shape, or file must change the gate in the same phase.

## Invariants enforced by tests

- the transitive closure is stdlib plus this module;
- the intra-module DAG is exactly `workflow -> action`, `transaction -> action`,
  and `resilience -> action`, with `plan` a leaf and `action` the sink;
- no `require`, no `replace`, no new package, no mutable global;
- no production goroutine, no call to `recover`, and no creating, deriving,
  shielding, or reading of context values (`Background`, `TODO`, `WithCancel`,
  `WithDeadline`, `WithValue`, `Value`, `WithoutCancel`, `AfterFunc`); the only
  exception is `context.WithTimeout` in the `resilience` package;
- `Pair` and `AttemptResult` carry exactly the decided fields and zero methods;
- every export has exercised consumption outside its own package, proven by
  `go/parser` without type information (`adr:workflow:ast-evidence-for-exercised-surface-gate`): an exact selector
  of import path and symbol, with aliases resolved; a dot import of a tracked
  package is rejected, a blank import is ignored, and shadowed identifiers are
  discarded; comments, strings, prefixes, and foreign packages do not count. A
  consumed function reaches only the local types of its results, and a consumed
  type, the types of its exported fields; constants require a direct selector.
  The consumers are the module's production code and the external
  `example_test.go` files, except `internal/archtest`. `go vet ./...` checks
  that the consumers compile; `go test ./...` runs the examples with declared
  output;
- production surfaces and files are exact allowlists; the required-documents
  list is checked for presence only, along with frontmatter keys and selected
  tokens;
- every gate carries negative evidence that it can fail.

## Execution characteristics

Construction and composition are lazy; `Run` is the only path to effects, and
it passes the `ctx` to the Action unchanged. There is no automatic preemption:
cancellation is cooperative and depends on the Action observing the context.
Actions are re-runnable, with no memoization and no runtime. One closure per
combinator is measured by benchmark, with no universal zero-allocation promise.
