---
type: tdd
id: tdd:workflow:plan-resilience
context: workflow
title: Contratos de plan e resilience
status: Aceito
tags: [workflow, plan, resilience, retry, timeout, context, contracts]
related: [adr:workflow:010, adr:workflow:011, adr:workflow:002, tdd:workflow:action, tdd:workflow:core, tdd:workflow:transaction]
date: 2026-10-23
---
# TDD: Contratos de `plan` e `resilience`

## Escopo técnico

Este desenho detalha os contratos de dois pacotes de produção do módulo:
`workflow/plan` (ADR-010) e `workflow/resilience` (ADR-011). Os três pacotes
restantes — `workflow`, `workflow/action` e `workflow/transaction` — são
cobertos por [`tdd:workflow:core`](core.md),
[`tdd:workflow:action`](action.md) e
[`tdd:workflow:transaction`](transaction.md).

Não existe SDD do módulo workflow: `docs/sdd/` não existe e nenhum documento
declara `sdd:workflow:*`. Este TDD é, portanto, ancorado diretamente nos ADRs e
não referencia um SDD.

```text
plan                     (folha: somente slices e strings)
workflow ──► action ◄── transaction
                 ▲
                 └────── resilience   (somente context, action e time)
```

`plan` não importa pacote do módulo; `action` é o sumidouro. `resilience`
depende exatamente de `context`, `action` e `time`.

## Contratos públicos

```go
// plan
type Kind uint8 // KindEmpty, KindStep, KindSequence, KindTransaction
type Plan struct{ /* campos não exportados */ }
func Step(name string) Plan
func Sequence(plans ...Plan) Plan
func Transaction(plans ...Plan) Plan
func (Plan) Kind() Kind
func (Plan) Name() string
func (Plan) Children() []Plan
func (Plan) String() string
func (Kind) String() string

// resilience
type Backoff func(attempt int) time.Duration
type RetryPolicy struct {
	Attempts int
	When     func(error) bool
	Backoff  Backoff
}
func Retry[T any](policy RetryPolicy, source action.Action[T]) action.Action[T]
func Timeout[T any](duration time.Duration, source action.Action[T]) action.Action[T]
```

A superfície exportada é fixada por `TestExactPublicSurface`.

## Invariantes de `plan`

Fixadas por `plan/plan_test.go`, `plan/example_test.go` e pelos gates de
`internal/archtest`:

- o valor zero de `Plan` é `KindEmpty`, sem nome e com `Children() == nil`;
- `Step("")` entra em pânico com `"workflow/plan: empty step name"`; nomes não
  vazios são preservados verbatim, inclusive espaços;
- `Sequence()` e `Sequence(empty)` reduzem a `KindEmpty`; `Sequence(x)` devolve
  `x`; duas ou mais entradas preservam ordem e aninhamento, sem achatar;
- `Transaction(...)` sempre preserva o nó, inclusive vazio e com um filho;
- construtores e `Children()` clonam o slice: mutar a entrada ou o resultado
  não altera o `Plan`;
- `Kind.String()` devolve `empty`, `step`, `sequence`, `transaction` e
  `unknown` para qualquer valor indefinido;
- `Plan.String()` é canônico e determinístico, por exemplo
  `sequence(a, b, transaction(c, d))` e `transaction()`, sem escape de nomes e
  sem promessa de round-trip; chamar `String()` não altera o `Plan`;
- leituras concorrentes de um mesmo `Plan` são seguras;
- `Plan` não exporta campos, expõe somente `Children`, `Kind`, `Name` e
  `String` (`TestPlanIsOpaqueAndNonExecutable`), não tem `Run` nem `Execute` e
  `plan` importa exatamente `slices` e `strings`: descrever não executa nem
  abre transação.

## Invariantes de `resilience`

Fixadas por `resilience/retry_test.go`, `retry_attempts_test.go` e
`timeout_test.go`.

### Construção

- `Retry` e `Timeout` são lazy: construir não chama a fonte;
- argumentos inválidos são erro de programação e entram em pânico na
  construção, com mensagens exatas e sem chamar fonte nem predicado:
  `"workflow/resilience: invalid attempts"` (`Attempts <= 0`),
  `"workflow/resilience: nil predicate"` (`When == nil`),
  `"workflow/resilience: nil action"` (fonte nula em `Retry` ou `Timeout`) e
  `"workflow/resilience: invalid timeout"` (duração `<= 0`);
- `Backoff` nulo é válido.

### `Retry`

- `Attempts` inclui a primeira chamada; a fonte roda no máximo `Attempts`
  vezes, parando no primeiro sucesso;
- `When` só é consultado depois de uma falha que ainda tem tentativa
  restante: com `Attempts` esgotado ele é chamado `Attempts-1` vezes; um
  predicado falso encerra sem backoff;
- `Backoff` recebe o número da tentativa que acabou de falhar, base 1
  (`[1 2]` para três tentativas); `Backoff` nulo e atraso zero tentam de
  imediato;
- atraso negativo entra em pânico com `"workflow/resilience: negative backoff"`
  depois de a fonte e o predicado rodarem uma vez;
- a ordem de decisão do laço é: (1) verificar `ctx.Err()`; (2) chamar a fonte;
  (3) sucesso devolve o valor; (4) com `Attempts` esgotado ou `When` falso, o
  erro da fonte é devolvido diretamente; (5) só então vem o backoff e a próxima
  tentativa. O erro da fonte sai sem embrulhar nem agregar e o valor é o zero
  de `T`; não há `errors.Join` entre tentativas nem entre o erro da fonte e o
  do contexto;
- o erro da fonte é o resultado somente quando a própria fonte encerra o laço
  (passo 4). Nesse caso o contexto não é consultado depois da chamada: um
  cancelamento ou prazo que ocorra durante ou depois da última tentativa, ou
  depois de `When` falso, não substitui o erro da fonte, e um sucesso da fonte
  continua sendo devolvido mesmo que o contexto tenha sido cancelado durante a
  chamada. O cancelamento do pai não sobrepõe um resultado terminal da fonte;
- `Retry` devolve exatamente `ctx.Err()` — `context.Canceled` ou
  `context.DeadlineExceeded`, sem embrulho — somente quando observa a parada do
  contexto pai em um dos dois pontos de verificação: na entrada de uma
  tentativa (passo 1, inclusive a primeira, sem chamar fonte, predicado nem
  backoff) ou na espera do backoff (passo 5). Depois de uma falha retentável com
  tentativa restante, um pai já parado é observado na espera quando há atraso
  positivo ou zero (atraso zero apenas consulta `ctx.Err()`) e, com `Backoff`
  nulo, na entrada da tentativa seguinte; em nenhum dos casos a fonte é chamada
  de novo. O erro retentável anterior é descartado nesse caso;
- pânicos da fonte, de `When` e de `Backoff` escapam inalterados;
- o mesmo `Retry` construído pode ser executado concorrentemente.

### `Timeout`

- cada execução deriva um contexto filho com prazo `duration` a partir do
  `ctx` recebido; valores do contexto pai continuam visíveis;
- a fonte recebe o filho, nunca o próprio pai, e seu valor e erro retornam sem
  alteração;
- um pai cancelado antes da execução chega cancelado à fonte;
- uma fonte que observa `ctx.Done()` e devolve `ctx.Err()` produz
  `context.DeadlineExceeded` no prazo;
- pânicos da fonte escapam inalterados e o mesmo `Timeout` suporta execução
  concorrente.

## Cancelamento cooperativo

`action.Run` repassa o `ctx` à Action sem alteração e sem verificá-lo
(`action/action.go`). Nada no módulo interrompe uma função em execução.
`TestTimeoutDoesNotPreemptIgnoringSource` prova que uma fonte que ignora o
contexto continua rodando depois do prazo e `Run` só retorna quando ela
retorna. O cancelamento é cooperativo: a fonte, ou o adaptador que ela chama,
precisa observar `ctx`. `Retry` é a única exceção parcial, porque verifica
`ctx.Err()` entre tentativas e durante o backoff, nunca dentro da fonte.

Não há goroutine de produção, `recover` nem preempção automática.

## Composição de orçamentos

Fixada por `TestTimeoutOutsideRetryUsesOneTotalBudget` e
`TestRetryOutsideTimeoutCreatesPerAttemptBudgets`:

- `Timeout(d, Retry(p, src))` cria um único prazo total para todas as
  tentativas e esperas; com um backoff maior que o prazo, a fonte roda uma vez
  e o resultado é `context.DeadlineExceeded`;
- `Retry(p, Timeout(d, src))` cria um filho novo por tentativa, cada um com
  prazo e contexto próprios.

A ordem é escolha explícita do chamador; o módulo não oferece política
combinada.

## Exceção de `context.WithTimeout`

O gate de contexto (`TestNoHiddenContextPolicy`) proíbe em produção
`Background`, `TODO`, `Value`, `WithValue`, `WithCancel`, `WithDeadline`,
`WithoutCancel`, `AfterFunc` e `WithTimeout`, incluindo as variantes `Cause`.
A única exceção é `context.WithTimeout` dentro do pacote `resilience`, usado
por `Timeout`. Nenhum outro pacote cria, cancela ou protege contexto.

## O que estes pacotes não fazem

- `Plan` não executa, não compila para `Action`, não registra nem agenda;
- `Retry` e `Timeout` não são motor de execução: não persistem estado nem
  agendam, e não incluem jitter, circuit breaker ou classificação de erros
  (ADR-011);
- `Retry` não torna uma operação idempotente; repetir efeitos exige que a
  operação, ou seu adaptador, seja segura para repetição.

## Nota sobre os gates de documentação

`TestDocumentationSpine` verifica apenas que este arquivo existe e que a
frontmatter tem as chaves obrigatórias, status `Aceito` e nenhum `deciders`.
O gate é de presença: ele não prova que a lista de documentos é uma allowlist
exata nem que o texto corresponde aos testes citados acima.
