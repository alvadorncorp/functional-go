---
type: tdd
id: tdd:modeling:reducer-replay
context: modeling
title: Modeling reducer and replay
status: Accepted
tags: [modeling, reducer, replay, iter, generics, laws, testing]
related: [adr:modeling:minimal-type-safe-domain-modeling, adr:modeling:root-facade-internal-decision, adr:modeling:change-capability-boundary, adr:modeling:policy-capability-boundary, adr:modeling:reducer-capability-boundary, adr:modeling:streaming-replay-project-seq, tdd:modeling:policy]
date: 2026-08-12
---
# TDD: Modeling reducer and replay

## Escopo técnico

Este desenho descreve o pacote importável
`go.alvadorn.org/functional/modeling/reducer`. Decision, Transition,
FallibleTransition, Invariant, Rule, Requirement, composição, métodos, caminhos
de import da raiz, `modeling/internal/decision`, `modeling/change` e
`modeling/policy` permanecem contratos separados.

Reducer modela transformação fato → estado. Distingue-se de Transition
(intenção/input → output), de Decision (próximo estado + eventos emitidos), de
Change (evolução de estado sem eventos) e de Policy (estratégia a partir de
contexto).

O parâmetro de eventos de `Replay` é `iter.Seq[E]` (`adr:modeling:streaming-replay-project-seq`). A escolha anterior
`immutable.List[E]` em `adr:modeling:reducer-capability-boundary` fica supersedida para esse parâmetro.

## Contratos públicos

O pacote exporta exatamente:

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S
```

Funções ordinárias `func(S, E) S` são atribuíveis e invocáveis sem wrappers.
Reducer devolve estado diretamente e não possui canal de erro. Replay aceita
somente `iter.Seq[E]`; não há sobrecargas de slice, List, Decision ou Apply.
Ambos os parâmetros de tipo permanecem `any` irrestrito, inclusive tipos de
estado e evento não comparáveis. A semântica é de valor: não há
`Reducer[*S, E]`, `Ref[T]`, `Box` ou `Readonly` no metamodelo.

## Semântica e leis de Replay

Replay é um fold esquerdo síncrono de **passagem única** sobre os fatos na
ordem de encontro da sequência:

```text
s1 = reduce(s0, e1)
s2 = reduce(s1, e2)
...
sn = reduce(s(n-1), en)
```

Invoca o redutor exatamente uma vez por evento; não reverte nem paraleliza.
Sequência vazia invoca o redutor zero vezes e devolve o estado inicial.

Reducer nil não é invocado na sequência vazia. Para entrada não vazia, vale o
comportamento ordinário de Go (panic em chamada de função nil); Replay não
adiciona guarda nem contrato de erro alternativo.

Lei de concatenação (quando a aplicação define `A ++ B` de forma observável):

```text
Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)
```

Replay repetido devolve estado igual quando o redutor é puro e determinístico.
Pureza é responsabilidade da aplicação; modeling não detecta relógio,
aleatoriedade, estado global, I/O ou efeitos colaterais.

Como a sequência é de passagem única, quem precisar de multi-pass materializa
explicitamente **antes** de chamar Replay.

## Representação de eventos e ponte com Decision

`E` permanece `any` irrestrito. A aplicação é dona de uniões heterogêneas,
switches e política de kind desconhecido. modeling não adiciona marker,
reflection, registry, handler map, schema/version nem hierarquia de eventos.

Decision e Reducer permanecem ortogonais. Quando os eventos da Decision
reconstróem estado, a ponte é direta:

```go
state := reducer.Replay(previous, decisionValue.Events(), reduce)
```

Não é necessário `immutable.CollectList`. A biblioteca não compara
`Decision.State()` com o resultado de Replay nem oferece
`ReplayDecision`/`VerifyDecision`; eventos de integração não são
universalmente reconstrutores de estado.

## Fronteira de implementação e dependências

Replay itera `for event := range events` e encadeia uma variável local de
estado. Não materializa `[]E`, não chama `Items()`/`CollectList` e não usa
`fold.FoldLeft`.

Imports de produção do pacote: somente a biblioteca padrão (`iter`). Não
importa `functional/*`, a raiz de modeling nem capability irmã.

Testes de arquitetura devem impor:

- exatamente dois exports, `Reducer` e `Replay`;
- assinaturas de compilação com `iter.Seq[E]`;
- ausência de dependência de sibling e de `functional/*` em produção;
- nome `Replay` somente em `modeling/reducer`;
- range direto sobre `events` (passagem única, sem materialização);
- sem aresta inversa functional→modeling, globals mutáveis de produção ou
  superfície mágica de runtime.

## Superfície excluída

Não há Apply, FallibleReducer, ReplaySlice, ReplayList, ReplaySeq (além do
próprio `iter.Seq` no parâmetro), Reduce, ReplayDecision, VerifyDecision,
marker Event/DomainEvent, reflection, registry, dispatcher, loader de stream,
Event Store, snapshot store, projection, serialização, migração de versão,
persistência, transporte, database, filesystem, HTTP, broker, replay baseado
em goroutine, atualização de dependência ou configuração global.

## Verificação

Executar testes focados de reducer e arquitetura (incluindo passagem única e
ausência de import functional), testes do módulo em workspace e
`GOWORK=off`, race, vet, tidy-diff, lint sem caps ocultos e listagem exata de
pacotes. Fechar com `make test-unit`, `make test-race`, `make lint`, `make vet`
e `git diff --check` no workspace quando a entrega incluir código. Expansão de
benchmark de CI, tags e publicação ficam fora deste TDD documental.
