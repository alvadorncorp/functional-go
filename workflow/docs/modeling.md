# modeling integration

```text
Workflow
   │
   ▼
load state          effectful capability
   │
   ▼
domain function     pure
   │
   ▼
modeling.Decision   value
   │
   ▼
transaction         effectful boundary
   ├── persist state
   └── persist resulting events
```

Ownership stays unambiguous:

```text
Specification, Policy, Rule, Requirement, Transition, Invariant  -> domain (via modeling)
Decision                                                          -> domain result
Load, Save                                                        -> workflow capability
Outbox                                                            -> application capability
Transaction                                                       -> workflow execution boundary
```

## Load, decide, confirm

```go
decision, err := domain.Withdraw(account, cmd)
```

The workflow calls an ordinary domain function. It does not use
`modeling.Requirement` directly, because the domain function is where the
rules belong together.

## Specification classifies, Policy chooses

```text
WithdrawCommand
      |
      v
RequiresManualReview Specification   pure classification
      |
      v
SelectWithdrawalMode Policy          pure strategy
      |
      v
Withdraw domain function             requirements + transition + invariant + decision
```

Specification and Policy receive only already-loaded data. Databases, HTTP,
brokers, clocks, the filesystem, and `context.Context` stay outside both. The
explicit closure between the named Specification and `policy.Choose` shows the
boundary without creating a dependency between the packages.

## Policy chooses, workflow executes

```text
Amount ──► WithdrawalPolicy ──┬──► Immediate        ──► persist withdrawal
                              └──► RequiresApproval ──► persist approval request
```

The threshold lives in the domain. A workflow that wrote
`if amount > 10_000` would be hiding a business rule inside orchestration. The
workflow only interprets the mode the domain chose.
