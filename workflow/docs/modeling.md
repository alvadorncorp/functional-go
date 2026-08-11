# Integração com modeling

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

A propriedade não se confunde:

```text
Rule, Requirement, Transition, Invariant, Policy  →  domain (via modeling)
Decision                                          →  domain result
Load, Save                                        →  workflow capability
Outbox                                            →  application capability
Transaction                                       →  workflow execution boundary
```

## Carregar, decidir, confirmar

```go
decision, err := domain.Withdraw(account, cmd)
```

O workflow chama uma função de domínio ordinária. Ele não usa
`modeling.Requirement` diretamente, porque a função de domínio é onde as regras
pertencem juntas.

## Policy escolhe, workflow executa

```text
Amount ──► WithdrawalPolicy ──┬──► Immediate        ──► persist withdrawal
                              └──► RequiresApproval ──► persist approval request
```

O limiar vive no domínio. Um workflow que escrevesse `if amount > 10_000` estaria
escondendo uma regra de negócio dentro da orquestração. O workflow apenas
interpreta o modo que o domínio escolheu.
