---
type: adr
id: adr:workflow:functional-dependency-injection-through-capabilities
context: workflow
title: Functional dependency injection through capabilities
status: Accepted
tags: [workflow, dependency-injection, capabilities, composition-root]
deciders: [alvadorn]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:explicit-transaction-scopes, tdd:workflow:composition]
date: 2026-08-10
---
# Functional dependency injection through capabilities

## Context and problem

Dependency injection in Go is usually solved with a container, a registry, or a service locator. Each moves wiring errors from compile time to run time and makes an operation's dependency graph unreadable from its signature.

## Decision drivers

- the dependency graph visible in the constructor's signature;
- test doubles without a mocking framework;
- compile-time detection of a missing dependency;
- no global mutable state.

## Decision

Dependencies are supplied as functions, closures, and small capability structs, assembled by the application's composition root. A capability is a struct of functions such as `Accounts{ Load func(AccountID) action.Action[Account] }`. The module defines no universal port: `Repository`, `Store`, `CommandBus`, and relatives are banned exported names.

## Alternatives considered

- **A container with `Register` / `Resolve`.** Rejected: runtime wiring failures, reflection, and global registration, all banned by name in the archtest.
- **Interfaces as the mandatory injection mechanism.** Rejected as *mandatory*, not as an option: an interface remains correct where there is genuine polymorphism. It is simply not required for injection.
- **Passing dependencies through `context.Context`.** Rejected: context carries lifecycle, not capabilities, and context values defeat compile-time checking.

## Consequences

Building a workflow with fake capabilities requires no framework: a struct literal with closures is the double. Because dependencies are captured by closure at construction, the graph is stable: no later reassignment changes what an already-built workflow uses.

## Non-goals

No `workflow/di` package, no annotations, no code generation, and no automatic wiring.
