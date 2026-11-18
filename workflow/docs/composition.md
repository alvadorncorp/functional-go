# Composition

Workflows build up a vocabulary:

```text
LoadUser  +  RequireActive        =  LoadActiveUser
LoadActiveUser  +  LoadPermissions =  AuthorizedUser
AuthorizedUser  +  ChangeEmail     =  ChangeEmailWorkflow
```

`LoadActiveUser` is an ordinary `Workflow`. Nothing marks it as composed, and
larger workflows cannot tell the difference, which is exactly why the
vocabulary can grow without consumers learning an internal tree.

```go
loadActiveUser := workflow.Then(loadUser, requireActive)
```

## Composition is binary

Go generics do not preserve an intermediate type that changes across a
variadic list, so `Then` takes two workflows. Chain them:

```go
ab  := workflow.Then(a, b)
abc := workflow.Then(ab, c)
```

## Not everything should be a workflow

A pure function stays a function. `func NormalizeEmail(string) string` gains
nothing by becoming a `Workflow`. Reach for one when there is an effect,
reusable orchestration, dependency capture, composition across capabilities,
or an execution lifecycle. Uniformity is not worth losing meaning.

Likewise, a pure domain function that returns a `modeling.Decision` should
stay a pure domain function; `MapFallible` exists so it can enter a workflow
without disguising itself as an effect.

## Adapting and supplying input

`Contramap(wf, mapper)` turns `Workflow[I,O]` into `Workflow[A,O]` through a
projection `A → I`. The mapper runs when the workflow is applied.
`Provide(wf, input)` eliminates the input and returns `Action[O]`; it does not
call `Run`.

## Dependency, independence, and parallelism

| Operation | Relation | Result |
|---|---|---|
| `FlatMap` | the second Action depends on the first value | `Action[B]` |
| `Zip` | independent results, sequential execution | `Action[Pair[A,B]]` |
| `Both` | independent results, sequential execution | `Action[C]` |
| `Traverse` | one Action per item, in index order, stops at the first error | `Action[[]B]` |

Independence describes data flow, not concurrency. Order remains left →
right, with no goroutines.
