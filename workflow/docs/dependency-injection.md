# Dependency injection

A capability is a small struct of functions:

```go
type Users struct {
    Load func(UserID) action.Action[User]
    Save func(User) action.Action[struct{}]
}
```

A workflow constructor takes what it needs and returns a value:

```go
func ChangeEmail(users Users, clock Clock) workflow.Workflow[ChangeEmailCommand, User]
```

The composition root is the only place that knows several concrete
implementations at once:

```text
PostgreSQL   SMTP   Clock
        │
        ▼
   capabilities
        │
        ▼
    workflows
        │
        ▼
   HTTP adapters
```

## Why this is not a service locator

Nothing is registered and nothing is resolved. A missing dependency is a
compile error, not a runtime panic. The dependency graph of an operation is
readable from the constructor's signature, and a workflow built with one set
of capabilities keeps using them; nothing resolves them late and swaps them
out.

## Not every dependency is an effect

A clock and an identifier generator need no context, so they are plain
functions:

```go
type Clock func() time.Time
type NewMessageID func() MessageID
```

Use the simplest form that keeps things testable. Wrapping `Clock` in an
`Action` would add ceremony and no information.

## Testing without a framework

A struct literal with closures is the test double:

```go
users := Users{
    Load: func(id UserID) action.Action[User] { return action.Pure(existing) },
    Save: func(u User) action.Action[struct{}] { saved = u; return action.Pure(struct{}{}) },
}
```

## `Provide` supplies input, not dependencies

`workflow.Provide(loadUser, id)` applies a known input to a workflow and
returns an Action. It does not register, look up, or replace capabilities.
Dependencies still enter explicitly through the constructor; `Inject`,
`Resolve`, `Container`, and `Registry` remain forbidden.
