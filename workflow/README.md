# workflow

`workflow` provides small, type-safe primitives for building, composing, and
running reusable workflows with effects in Go.

Workflows describe application orchestration; domain rules stay in the domain,
and infrastructure stays behind capabilities the application defines.

```go
import (
    "go.alvadorn.org/functional/workflow"
    "go.alvadorn.org/functional/workflow/action"
)

type Users struct {
    Find func(UserID) action.Action[User]
}

func LoadUser(users Users) workflow.Workflow[UserID, User] {
    return workflow.From(users.Find)
}

loadUser := LoadUser(users)

program := loadUser(id)
// No effect has executed yet.

user, err := action.Run(ctx, program)
```

## The three moments

Dependencies enter at construction, input enters at application, and context
enters at execution. Building and composing workflows performs no I/O; only
`action.Run` starts anything.

## Packages

| Import path | Contents |
|---|---|
| `go.alvadorn.org/functional/workflow` | `Workflow[I,O]`, `From`, `Map`, `MapFallible`, `Then`, `Provide`, `Contramap` |
| `go.alvadorn.org/functional/workflow/action` | `Action[T]`, construction, transformation, sequential combination, observation, finalization, error reification, and resource safety |
| `go.alvadorn.org/functional/workflow/plan` | `plan.Plan` — an explicit, immutable, non-executable structural description |
| `go.alvadorn.org/functional/workflow/resilience` | `Retry`, `Timeout`, and backoff policy as explicit Action transformations |
| `go.alvadorn.org/functional/workflow/transaction` | `Boundary[S]`, `Within` |

Plan describes; Workflow builds; Action defers; Resilience transforms;
Boundary delimits; Run executes.

The core depends only on the standard library, and an architecture test keeps
it that way.

## Composition algebra

`Contramap` adapts input and `Provide` fixes the input without running
effects. `FlatMap` expresses dependency; `Zip` keeps two independent results
and `Both` combines them, and `Traverse` runs one Action per slice item,
always in sequential order. `Tap` observes success, `Attempt` turns an error
into data, `Ensure` finalizes normal returns, and `Bracket` scopes
acquire/use/release without owning the concrete resource.

## What it is not

It is not a workflow engine, a DDD framework, a dependency-injection
container, or an effect system. It promises no durability, no parallelism, and
no distributed atomicity. Retry and timeout exist only as explicit,
cooperative transformations in `workflow/resilience`.

## Documentation

Start with the [overview](docs/overview.md), then the
[architecture](docs/architecture.md). The full index is in
[docs/README.md](docs/README.md), including [Plan](docs/plan.md),
[Resilience](docs/resilience.md), ADRs, and technical designs.

The [runnable examples](docs/examples.md) use only the public API and
in-memory adapters.
