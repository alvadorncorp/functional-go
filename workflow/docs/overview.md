# Visão geral

Três módulos, três perguntas:

| Módulo | Responsabilidade | Pergunta que responde | Exemplo |
|---|---|---|---|
| `functional` | mecânicas funcionais genéricas | como os valores se compõem? | `Transform`, `Optional`, `Result`, coleções imutáveis |
| `modeling` | semântica de domínio pura | o que o domínio decide? | `Requirement`, `Transition`, `Invariant`, `Decision`, `Policy` |
| `workflow` | composição reutilizável com efeitos | como a operação é conduzida? | `Workflow`, `Action`, `Boundary`, `Plan`, `Retry`, `Timeout` |
| aplicação | regras e capacidades concretas | o que é este negócio? | `Withdraw`, `Accounts`, `Outbox` |
| adaptadores | execução técnica | como o efeito de fato acontece? | PostgreSQL, HTTP, Kafka |

## Descrição e execução

```text
Plan (explicit data; optional)       dependencies
          │                              │ construction
          │ describes                    ▼
          │                        Workflow[I,O]
          │                              │ explicit input I
          │                              ▼
          │                         Action[O]
          │                              │ optional resilience
          │                              ▼
          └─────────────────────── tooling only
                                         │ Run(context)
                                         ▼
                                   result / error
```

A construção é livre de contexto; a execução é ligada a contexto. Construir um
workflow não realiza I/O, compor workflows não realiza I/O e aplicar um workflow
a uma entrada não realiza I/O. Apenas `action.Run` inicia qualquer coisa. Plan
é dado explícito e opcional: descreve estrutura para humanos e tooling, sem
executar.

## Glossário

**Workflow** — uma operação reutilizável, `I → Action[O]`.

**Action** — uma computação adiada e ligada a contexto, `Context → (O, error)`.
Reexecutável, nunca memoizada, nunca iniciada em segundo plano.

**Plan** — descrição estrutural explícita, imutável e não executável da
operação. Inspecionável sem contexto, sem I/O e sem abrir transação.

**Resilience** — transformações explícitas e cooperativas de `Action`
(`Retry`, `Timeout`) aplicadas pelo adapter que conhece o orçamento e a classe
de falha.

**Run** — o único ponto em que efeitos começam: `action.Run(ctx, program)`.

**Capacidade** — uma função, ou um pequeno struct de funções, fornecida a um
workflow para que ele execute uma operação externa.

**Boundary** — o escopo de transação concreto fornecido pelo adapter
(`transaction.Boundary[S]`), não um motor de atomicidade do núcleo.

**Decisão de domínio** — o resultado puro que o domínio produz, em geral um
`modeling.Decision`.

**Adaptador** — uma implementação concreta de capacidade usando uma tecnologia
externa.

Receitas de aplicação (handle COW, load-decide-confirm, outbox, DAG, ACL) estão em [patterns/](patterns/README.md). Um módulo DDD de runtime está fora de escopo ([no ddd runtime module](adr/workflow/no-ddd-runtime-module.md)).

## Dados de negócio e contexto de execução

Se uma informação pode alterar uma decisão de domínio — CustomerID, TenantID,
ActorID, quantidade — ela entra no input explícito do Workflow. Context é
reservado a lifetime operacional: cancelamento, deadline e propagação de
instrumentação feita pelos adapters. Não existe WorkflowContext,
ExecutionContext, RequestContext nem leitura de ctx.Value para regra de negócio.

## O que este módulo não é

`Workflow[I,O]` não implica persistência de execução, resume automático, timers
duráveis, agendamento distribuído, histórico, replay determinístico, leasing de
workers, checkpoints nem identificadores de workflow. Não compete com Temporal,
Cadence ou Camunda. A durabilidade de efeitos de negócio vem de mecanismos
explícitos — um outbox transacional, estado persistido, event store, fila,
agendador de jobs — que um workflow pode coordenar, mas não possui.

Tampouco é uma máquina de estados. Um workflow pode coordenar uma máquina de
estados de domínio, mas não tem ciclo de vida persistido próprio e não define
vocabulário `Pending`/`Running`/`Completed`.

## Dimensões de composição

| Dimensão | Operações | Significado |
|---|---|---|
| input | `Contramap`, `Provide` | adaptar e depois fornecer a entrada |
| computação | `FlatMap`, `Zip`, `Both`, `Traverse` | dependência ou independência sequencial |
| observação/finalização | `Tap`, `Attempt`, `Ensure` | sucesso, erro como dado e término normal |
| resource lifecycle | `Bracket`, `AcquireUseRelease` | acquire/use/release explícito |
| operacional | `resilience.Retry`, `resilience.Timeout` | orçamento e repetição seletiva, dono = camada que conhece o requisito operacional |

Essas dimensões descrevem efeitos; adapters continuam donos de conexões, leases,
locks, spans e qualquer política operacional concreta.
