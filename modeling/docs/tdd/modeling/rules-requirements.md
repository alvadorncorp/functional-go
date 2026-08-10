---
type: tdd
id: tdd:modeling:rules-requirements
context: modeling
title: Modeling rules and requirements
status: Aceito
tags: [modeling, rule, requirement, composition, generics, testing]
related: [adr:modeling:001, adr:modeling:003, tdd:modeling:core, tdd:modeling:decision-composition]
date: 2026-08-10
---
# TDD: Modeling rules and requirements

## Technical scope

The root `modeling` package remains the only externally importable product
package. Rule and Requirement are small nominal function types implemented in
root files. `modeling/internal/decision` stays Decision-only;
`modeling/internal/archtest` enforces dependencies and the public envelope.

## Public contracts

Rules and Requirements add exactly:

```go
type Rule[T any] func(T) error
func AllRules[T any](rules ...Rule[T]) Rule[T]
func AnyRule[T any](first Rule[T], rest ...Rule[T]) Rule[T]

type Requirement[S, I any] func(S, I) error
func AllRequirements[S, I any](requirements ...Requirement[S, I]) Requirement[S, I]
func AnyRequirement[S, I any](first Requirement[S, I], rest ...Requirement[S, I]) Requirement[S, I]
```

Ordinary functions with matching signatures are assignable without wrappers.
Rule is not an alias of Invariant, and Requirement is not an automatic
Transition pre-hook.

## Composition semantics

| Operation | Empty input | Success | Failure | Order |
|---|---|---|---|---|
| AllRules | nil | after every Rule returns nil | first error unchanged, then stop | declaration order |
| AllRequirements | nil | after every Requirement returns nil | first error unchanged, then stop | declaration order |
| AnyRule | excluded by first-operand signature | first nil, then stop | final evaluated error unchanged | declaration order |
| AnyRequirement | excluded by first-operand signature | first nil, then stop | final evaluated error unchanged | declaration order |

Every constituent receives the original arguments unchanged. A nil Rule or
Requirement is rejected while the composition is built, with exact panic text
`modeling: nil rule` or `modeling: nil requirement`. Returned errors remain
owned and interpreted by the application.

## Explicit domain flow

```text
Requirement[S,I]  S × I → error
        ↓ valid
Transition[S,I,O] S × I → O
        ↓
Invariant[O]      O → error
        ↓ valid
Decision[O,E]     O + E*
```

The example application invokes each piece explicitly. modeling supplies no
pipeline, runner, command handler, registry, rules engine, persistence, or
transport.

## Compatibility and architecture gates

Decision, Transition, FallibleTransition, Invariant, FailFast, and Accumulate
retain their exact contracts. No dependency or package is added. External tests
assert ordinary-function adaptation, call order/count, exact error identity,
empty All behavior, non-empty Any behavior, and nil rejection. An AST test caps
the exports of `rule.go` and `requirement.go`; the existing dependency,
public-signature, reverse-edge, runtime-name, global-state, and package gates
continue to inspect all production code.

## Excluded surface

There is no NotRule, NotRequirement, FirstFailure, RuleFromPredicate,
AccumulateRules, AccumulateRequirements, fluent method, Policy, Specification,
RuleEngine, RuleRegistry, runtime RuleSet, runner, executor, automatic pipeline,
new error hierarchy, public subpackage, reflection, annotation, or code generation.

## Verification

Run focused Rule/Requirement tests, then module test, race, vet, tidy-diff,
uncapped lint, and exact package listing. Finish with workspace `make test-unit`,
`make test-race`, `make lint`, `make vet`, and `git diff --check`. No database
integration, migration, benchmark expansion, dependency update, tag, or release
execution belongs to this design.
