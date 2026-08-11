# Visão geral

Três módulos, três perguntas:

| Módulo | Responsabilidade | Pergunta que responde | Exemplo |
|---|---|---|---|
| `functional` | mecânicas funcionais genéricas | como os valores se compõem? | `Transform`, `Optional`, `Result`, coleções imutáveis |
| `modeling` | semântica de domínio pura | o que o domínio decide? | `Requirement`, `Transition`, `Invariant`, `Decision`, `Policy` |
| `workflow` | composição reutilizável com efeitos | como a operação é conduzida? | `Workflow`, `Action`, fronteiras de transação |
| aplicação | regras e capacidades concretas | o que é este negócio? | `Withdraw`, `Accounts`, `Outbox` |
| adaptadores | execução técnica | como o efeito de fato acontece? | PostgreSQL, HTTP, Kafka |

## Os três momentos

```text
dependencies
     ↓  construction
Workflow[I,O]
     ↓  input
Action[O]
     ↓  context
Run
     ↓
result / error
```

A construção é livre de contexto; a execução é ligada a contexto. Construir um
workflow não realiza I/O, compor workflows não realiza I/O e aplicar um workflow
a uma entrada não realiza I/O. Apenas `action.Run` inicia qualquer coisa.

## Glossário

**Workflow** — uma operação reutilizável, `I → Action[O]`.

**Action** — uma computação adiada e ligada a contexto, `Context → (O, error)`.
Reexecutável, nunca memoizada, nunca iniciada em segundo plano.

**Capacidade** — uma função, ou um pequeno struct de funções, fornecida a um
workflow para que ele execute uma operação externa.

**Escopo de transação** — o conjunto de capacidades ligadas a uma fronteira de
transação concreta.

**Decisão de domínio** — o resultado puro que o domínio produz, em geral um
`modeling.Decision`.

**Adaptador** — uma implementação concreta de capacidade usando uma tecnologia
externa.

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

## Quatro dimensões de composição

| Dimensão | Operações | Significado |
|---|---|---|
| input | `Contramap`, `Provide` | adaptar e depois fornecer a entrada |
| computação | `FlatMap`, `Zip`, `Both` | dependência ou independência sequencial |
| observação/finalização | `Tap`, `Attempt`, `Ensure` | sucesso, erro como dado e término normal |
| resource lifecycle | `Bracket`, `AcquireUseRelease` | acquire/use/release explícito |

Essas dimensões descrevem efeitos; adapters continuam donos de conexões, leases,
locks, spans e qualquer política operacional concreta.
