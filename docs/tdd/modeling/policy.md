---
type: tdd
id: tdd:modeling:policy
context: modeling
title: Modeling Policy
status: Accepted
tags: [modeling, policy, generics, optional, predicate, testing]
related: [adr:modeling:policy-capability-boundary, adr:modeling:policy-composition-algebra, adr:modeling:specification-capability-boundary, tdd:modeling:specification, tdd:modeling:rules-requirements]
date: 2026-08-10
---
# TDD: Modeling Policy

## Technical scope

Create the public package `modeling/policy`, independent of the root and of
other capabilities. Change/Reducer/Replay are neither prerequisites nor part of
this delivery.

## Public contracts

```go
type Policy[I, O any] func(I) O
func Constant[I, O any](output O) Policy[I, O]
func Map[I, O, P any](p Policy[I, O], mapper func(O) P) Policy[I, P]
func Contramap[A, I, O any](p Policy[I, O], mapper func(A) I) Policy[A, O]
func When[I, O any](condition predicate.Predicate[I], p Policy[I, O]) func(I) optional.Optional[O]
func FirstApplicable[I, O any](first func(I) optional.Optional[O], rest ...func(I) optional.Optional[O]) func(I) optional.Optional[O]
func Fallback[I, O any](conditional func(I) optional.Optional[O], fallback Policy[I, O]) Policy[I, O]
func Compose[A, B, C any](first Policy[A, B], second Policy[B, C]) Policy[A, C]
func Zip[I, A, B, O any](left Policy[I, A], right Policy[I, B], combine func(A, B) O) Policy[I, O]
func Choose[I, O any](condition predicate.Predicate[I], whenTrue Policy[I, O], whenFalse Policy[I, O]) Policy[I, O]
func OrElse[I, O any](left func(I) optional.Optional[O], right func(I) optional.Optional[O]) func(I) optional.Optional[O]
```

Apply, FlatMap, a nominal Conditional, multi-arity input, methods, aliases, a
public tuple, and bridges are not part of the contract. Multiple data items
enter through an explicit context in I.

## Semantics

Constant ignores the input. Map transforms only the output and satisfies
identity and composition. Contramap adapts only the input and satisfies the
corresponding laws. When evaluates the predicate once and calls the policy only
on true; true always produces Some, including for a zero value. FirstApplicable
is non-empty, ordered, and stops at the first Some. Fallback preserves Some and
calls the fallback only on None. Compose evaluates first and then second. Zip
evaluates left, right, and combine in that order. Choose evaluates the
predicate once and only the selected branch. OrElse is binary, left-to-right,
and short-circuits; FirstApplicable remains the N-ary form.

Construction with a nil policy, mapper, predicate, or conditional panics with,
respectively, `modeling/policy: nil policy`, `modeling/policy: nil mapper`,
`modeling/policy: nil predicate`, and `modeling/policy: nil conditional`.
Zip also rejects a nil combiner with `modeling/policy: nil combiner`.

## Dependencies and architecture

Only `conditional.go` imports Predicate and Optional, besides the stdlib
`slices`. `modeling/policy` imports neither the root, nor siblings, nor
internal. AST gates limit the exact exports; dependency gates keep the
functional exception restricted to the package and its two functional imports.

## Tests and laws

Black-box tests pin signatures, ordinary-function assignability, zero values,
call count/order, Map/Contramap laws, short-circuit, Some(zero), all-None, lazy
fallback, Compose associativity, Zip order, the single Choose branch, OrElse
short-circuit, and nil panics. Executable examples show every combinator
without effects; bounded fuzzing covers Compose associativity.

## Compatibility

The existing public API stays intact. Tests run in the workspace and with
`GOWORK=off` against functional v1.1.0. The only metadata change is the
authorized functional bump to v1.1.0; there is no go.work change, migration,
config, benchmark, or publication.

## Excluded surface

There is no Rule/Requirement/Transition/Change/Reducer bridge, engine, registry,
dispatcher, reflection, effect, persistence, transport, applicability error,
concrete domain type, or Policy symbol in the root.

## Verification

Run focused tests, module tests, the published dependency check, race, vet,
tidy-diff, uncapped lint, package listing, and the workspace gates defined in
the plan.
