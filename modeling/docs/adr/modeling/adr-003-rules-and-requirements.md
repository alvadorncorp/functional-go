---
type: adr
id: adr:modeling:003
context: modeling
title: Nominal rules and operation requirements
status: Aceita
tags: [modeling, rule, requirement, composition, generics]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:002, tdd:modeling:v0-3-rules-requirements]
date: 2026-08-10
---
# ADR-003: Nominal rules and operation requirements

## Context and problem

Applications need to ask whether a value satisfies a domain condition and
whether an input may be applied to a current state before a Transition runs.
Invariant already answers a different question: whether a resulting state is
valid. Reusing one function representation for all three meanings would erase
domain intent, while a rules engine or automatic decision pipeline would move
application orchestration into a library runtime.

## Decision drivers

- keep Rule, Requirement, Transition, and Invariant semantically distinct;
- keep ordinary Go functions and application-owned errors ergonomic;
- preserve declaration order and exact error identity;
- give AND an empty identity without giving OR an artificial error identity;
- add no runtime, registry, dependency, public subpackage, or infrastructure.

## Options considered

### Alias Rule to Invariant

Rejected because identical representation does not mean identical domain
semantics. Aliasing would make operation policy and resulting-state validity
indistinguishable in public code.

### Reuse Boolean Predicate composition

Rejected because Predicate returns bool while Rule and Requirement preserve an
application-owned explanatory error. Converting through Predicate would discard
error identity or require a new modeling error abstraction.

### Make Any fully variadic and define a zero-input result

Rejected because nil is not a truthful OR identity and panic or ErrNoRules
would exist only to close the algebra at runtime.

### Require one Any operand and keep orchestration explicit (chosen)

Chosen because Go can enforce non-empty Any calls in the signature while the
application continues to invoke Requirement, Transition, Invariant, and
Decision explicitly.

## Decision

The root package adds exactly:

```go
type Rule[T any] func(T) error
func AllRules[T any](...Rule[T]) Rule[T]
func AnyRule[T any](Rule[T], ...Rule[T]) Rule[T]

type Requirement[S, I any] func(S, I) error
func AllRequirements[S, I any](...Requirement[S, I]) Requirement[S, I]
func AnyRequirement[S, I any](Requirement[S, I], ...Requirement[S, I]) Requirement[S, I]
```

All composition evaluates left-to-right, returns the first failure unchanged,
and accepts zero operands as success. Any composition requires a first operand,
returns nil at the first success, and returns the final evaluated error unchanged
when all operands fail. Nil constituent functions panic during composition with
`modeling: nil rule` or `modeling: nil requirement`, matching the existing
Invariant programmer-error convention.

Rule is a condition over one value. Requirement is an operation precondition
over state plus input. Transition computes the next state. Invariant validates
the result. The application owns domain errors and this invocation sequence.

## Consequences

Consumers gain a small additive algebra with no new import route or dependency.
Existing Decision, Transition, and Invariant contracts remain unchanged. The
module does not add negation, predicate bridges, accumulated violations,
fluent methods, engines, registries, runners, handlers, or automatic pipelines.

Before consumer adoption, rollback is a revert of the v0.3 commits. After
adoption, removing or changing any new exported signature or behavior is a
breaking public-contract decision requiring separate approval and release work.
