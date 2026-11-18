---
type: adr
id: adr:workflow:functional-integration-without-effect-duplication
context: workflow
title: Functional integration without effect duplication
status: Accepted
tags: [workflow, functional, dependencies, values, boundary]
deciders: [alvadorn]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:modeling-integration-boundary, tdd:workflow:core]
date: 2026-08-10
---
# Functional integration without effect duplication

## Context and problem

`functional` already provides `Optional`, `Result`, `Either`, `Validated`, `Transform`, `Predicate`, immutable collections, and `state`. A new functional-like module risks duplicating them or forcing every consumer through them.

## Decision drivers

- one owner per concept;
- no ceremony imposed on trivial operations;
- a core whose dependency closure stays minimal.

## Decision

The `workflow` core is stdlib only: `workflow/go.mod` declares no `require` and no `replace`, enforced by a test that parses the file named by `go env GOMOD`. The module introduces only what needs execution lifecycle (`Action`, `Workflow`, and transaction boundaries) and duplicates nothing from `functional`.

`Action` is deliberately not moved into `functional`: it carries `context.Context` and the application's execution semantics, while `functional` stays generic and free of any effect system.

## Alternatives considered

- **Returning `result.Result[T, E]` from `Run`.** Rejected: `(T, error)` is what every Go caller, adapter, and linter already understands, and `Result` remains available to applications that want it.
- **Automatic `FromResult` / `ToResult` bridges.** Rejected: `functional` itself avoids universal bridges between its value families, and reintroducing that coupling sideways would be worse.

## Consequences

A consumer can use `workflow` without importing `functional`. Comparing the neighboring concepts keeps them distinct: `functional.Transform[T]` is `T -> T`, `modeling.Change[S]` is `S -> S`, `workflow.Workflow[I,O]` is `I -> Action[O]`, and `action.Action[O]` is `Context -> (O, error)`.

## Non-goals

No monad family, no higher-kinded emulation, and no `Optional`/`Result`/`Either` equivalents inside `workflow`.
