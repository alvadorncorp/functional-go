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
Specification, Policy, Rule, Requirement, Transition, Invariant  -> domain (via modeling)
Decision                                                          -> domain result
Load, Save                                                        -> workflow capability
Outbox                                                            -> application capability
Transaction                                                       -> workflow execution boundary
```

## Carregar, decidir, confirmar

```go
decision, err := domain.Withdraw(account, cmd)
```

O workflow chama uma função de domínio ordinária. Ele não usa
`modeling.Requirement` diretamente, porque a função de domínio é onde as regras
pertencem juntas.

## Specification classifica, Policy escolhe

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

Specification e Policy recebem somente dados já carregados. Banco, HTTP,
broker, clock, filesystem e `context.Context` permanecem fora das duas. A
closure explícita entre a Specification nominal e `policy.Choose` mostra a
fronteira sem criar dependência entre os packages.

## Policy escolhe, workflow executa

```text
Amount ──► WithdrawalPolicy ──┬──► Immediate        ──► persist withdrawal
                              └──► RequiresApproval ──► persist approval request
```

O limiar vive no domínio. Um workflow que escrevesse `if amount > 10_000` estaria
escondendo uma regra de negócio dentro da orquestração. O workflow apenas
interpreta o modo que o domínio escolheu.
