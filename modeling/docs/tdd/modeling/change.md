---
type: tdd
id: tdd:modeling:change
context: modeling
title: Modeling Change and state evolution
status: Aceito
tags: [modeling, change, state, composition, endomorphism, testing]
related: [adr:modeling:001, adr:modeling:004, tdd:modeling:rules-requirements]
date: 2026-08-10
---
# TDD: Modeling Change and state evolution

## Technical scope

`modeling/change` is a new externally importable leaf inside the existing
modeling module. It does not re-export from the root or use
`modeling/internal/decision`. Internally it converts to released
`functional/primitives/function.Transform` values while keeping every public
signature modeling-owned.

When this contract was written the repository had no Reducer/Replay
implementation or release. Reducer/Replay was introduced later as an
independent capability (`tdd:modeling:reducer-replay`, ADR-006); ADR-010 only
partially amended it, moving `Replay` events to `iter.Seq`. Change has no
dependency on those concepts. This technical contract authorizes only the
Change capability; tagging and release numbering remain separate.

## Public contracts

Change adds exactly:

```go
type Change[S any] func(S) S
func Identity[S any]() Change[S]
func Compose[S any](changes ...Change[S]) Change[S]
func Pipe[S any](changes ...Change[S]) Change[S]
func Apply[S any](state S, transformation Change[S]) S
func ApplyAll[S any](state S, changes ...Change[S]) S
```

Ordinary functions with signature `func(S) S` are assignable to `Change[S]`
without constructors. Generic type identity prevents composition across
different state types.

## Evaluation semantics

| Operation | Empty input | Order | Result |
|---|---|---|---|
| Identity | n/a | n/a | input state unchanged |
| Compose(f,g,h) | Identity | right-to-left | `f(g(h(state)))` |
| Pipe(f,g,h) | Identity | left-to-right | `h(g(f(state)))` |
| Apply(state,f) | n/a | one direct call | `f(state)` |
| ApplyAll(state,f,g,h) | state unchanged | left-to-right | `h(g(f(state)))` |

Compose and Pipe are closed over `Change[S]`. Both satisfy left and right
identity and associativity. ApplyAll is observationally equal to
`Pipe(changes...)(state)`.

## Nil and immutability

Identity is the supported representation of no change. Nil is not a sentinel,
is not skipped, and receives no eager validation or custom panic contract;
invocation retains ordinary Go nil-function behavior.

Change is conceptually immutable, but the type system cannot enforce deep
immutability. A struct value may be copied and updated locally. Implementations
that touch slices, maps, pointers, or other references own the copy/aliasing
decision. The package uses no reflection or deep-copy machinery.

## Concept boundaries

- Transition receives state plus explicit input and may return another output.
- Change has already captured its context and always maps S to S.
- Decision adds events; Change emits none.
- Invariant validates a result; Change validates nothing.
- Rule and Requirement decide applicability; Change has no guard.
- Reducer-like functions may be understood as selecting a Change from an event,
  but this contract adds no Reducer API or dependency.
- functional Transform is structurally similar and semantically independent;
  Change privately adapts its released composition implementation.
- Change is not a Patch: it carries behavior, not a serializable structural diff.

## Package and file envelope

Production files are exactly `change.go`, `composition.go`, and `apply.go`.
Black-box contract files are exactly `change_test.go`, `composition_test.go`,
and `apply_test.go`. Package documentation lives in `change.go`; the executable
domain example lives in `change_test.go`.

## Compatibility and architecture gates

All existing root modeling signatures remain unchanged. The new child package
uses only the approved functional function path; the existing requirement moves
to v1.1.0. It adds no module path, workspace change, migration, configuration,
or reverse edge. Architecture tests cap the three production files, three test files, and six exports,
require package enumeration, and continue to reject forbidden imports,
functional public types, runtime names, mutable globals, and functional-to-
modeling dependencies.

Before consumer adoption, rollback is a commit revert. After publication,
incompatible removal or behavior changes require a separately approved release.

## Excluded surface

There is no root Change alias, public Transform alias, duplicated composition loop, ChangeTransition, Change-backed
Reducer, event emission, validation, guarding, ConditionalChange,
ValidatedChange, Patch, Diff, state manager, runtime, registry, runner, builder,
fluent method, interface, reflection, deep copy, code generation, benchmark,
fuzz target, additional module dependency, migration, global config, tag, or publication.

## Verification

Run focused Change and architecture tests, complete module tests with and
without the workspace, race, vet, tidy-diff, exact package listing, and uncapped
lint. Finish with workspace `make test-unit`, `make test-race`, `make lint`,
a fail-fast per-module `go vet` loop, `git diff --check`, and a fresh
`gopher:review --mode full`.
