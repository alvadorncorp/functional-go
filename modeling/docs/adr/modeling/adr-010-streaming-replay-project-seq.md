---
type: adr
id: adr:modeling:010
context: modeling
title: Replay e Project sobre iter.Seq com dependência somente stdlib
status: Aceita
tags: [modeling, reducer, projection, iter, streaming, compatibility, stdlib]
deciders: [alvadorn]
related: [adr:modeling:006, adr:modeling:007, tdd:modeling:reducer-replay, tdd:modeling:projection]
date: 2026-08-12
---
# ADR-010: Replay e Project sobre `iter.Seq` com dependência somente stdlib

## Contexto e problema

ADR-006 e ADR-007 estabeleceram as capabilities `modeling/reducer` e
`modeling/projection` com fold síncrono sobre fatos já ocorridos. Ambas
aceitaram `immutable.List[E]` no parâmetro `events` e importaram
`functional/immutable` apenas para esse contrato.

`Decision[S,E].Events()` já expõe `iter.Seq[E]`. Para reaplicar ou projetar
eventos de uma Decision, a aplicação precisava materializar a sequência com
`immutable.CollectList` antes de chamar `Replay`/`Project`. Esse passo:

- aloca uma lista intermediária desnecessária para um fold de passagem única;
- acopla reducer e projection a `functional/immutable` sem ganho semântico;
- impede o uso direto de qualquer `iter.Seq[E]` (streams lazy, concatenação
  sob demanda, geradores de teste) sem materialização prévia.

O código já enviado adota `iter.Seq[E]` e remove a dependência de produção de
`functional/immutable` em ambos os pacotes. Esta ADR registra a decisão de
contrato público e a supersessão parcial de ADR-006/007 quanto ao parâmetro
`events`.

## Direcionadores

- preservar a semântica de fold esquerdo síncrono e a ordem de encontro;
- eliminar materialização intermediária em `Replay` e `Project`;
- permitir `reducer.Replay(initial, decision.Events(), reduce)` e o análogo
  em projection sem `CollectList`;
- manter `Reducer` e `Projection` como `func` com semântica de valor;
- restringir imports de produção de reducer/projection ao stdlib (`iter`);
- não enfraquecer lint (`hugeParam` e demais regras) nem introduzir
  ponteiros no metamodelo para contornar o custo de cópia de valores grandes;
- documentar a quebra de contrato público já presente nas tags `modeling`
  até `v0.6.0`.

## Alternativas consideradas

### Manter `immutable.List[E]` (status quo de ADR-006/007)

Rejeitada. Obriga materialização quando a fonte já é `iter.Seq[E]` e mantém
acoplamento desnecessário a `functional/immutable`.

### Aceitar `[]E` ou sobrecargas List + Seq

Rejeitada. Multiplica superfície pública, favorece cópia para slice e dilui o
contrato único de passagem única.

### `fold.FoldLeft` sobre slice materializado

Rejeitada. `FoldLeft` opera em `[]T` e exige `Items()` ou conversão equivalente,
reintroduzindo alocação e dependência que esta decisão elimina.

### `Reducer`/`Projection` como ponteiro ou `immutable.Ref[T]` / `Box` / `Readonly`

Rejeitada. O metamodelo permanece com funções e valores; o custo de estados
grandes é problema da aplicação (representação própria), não motivo para
ponteirizar o tipo de capability. `hugeParam` continua como guarda-corpo de
lint sobre o tamanho de valores concretos, não como justificativa para
alterar a forma do metamodelo nem para `//nolint`.

### Desabilitar ou suprimir `hugeParam` / regras de lint

Rejeitada. Lint permanece efetivo; a correção de valores grandes ocorre na
modelagem da aplicação, não por supressão no pacote.

### `iter.Seq[E]` com fold de passagem única (escolhida)

Escolhida: um único parâmetro de sequência, consumo exatamente uma vez, zero
coleção intermediária e ponte direto com `Decision.Events()`.

## Decisão

Os contratos públicos permanecem com exatamente dois exports por pacote, com
o parâmetro de eventos em `iter.Seq[E]`:

```go
// modeling/reducer
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S

// modeling/projection
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Semântica:

- fold esquerdo síncrono de passagem única: cada evento é consumido exatamente
  uma vez, na ordem de encontro da sequência;
- sequência vazia devolve `initial` sem invocar o redutor/projeção;
- não há materialização de `[]E`, `immutable.List[E]` nem outra coleção
  intermediária dentro de `Replay`/`Project`;
- quem precisar de multi-pass materializa explicitamente **antes** de chamar
  a API;
- ponte Decision: `Replay(s, decision.Events(), r)` e
  `Project(p, decision.Events(), f)` sem `CollectList`.

Dependência de produção:

```text
application → modeling/reducer    → stdlib (iter)
application → modeling/projection → stdlib (iter)
```

Reducer e projection não importam `functional/*`, o root de modeling nem
capabilities irmãs. A direção `application → modeling root →
modeling/internal/decision → functional/immutable` permanece para Decision;
somente o parâmetro de eventos de Replay/Project deixa de exigir List.

Esta ADR **supersede** a escolha de `immutable.List[E]` em ADR-006 e ADR-007
para o parâmetro `events`. Permanecem em vigor: boundary dos pacotes,
identidade nominal de `Reducer`/`Projection`, exclusões de runtime ES/CQRS,
ausência de reexport na raiz e independência entre as duas capabilities.

## Compatibilidade e rollback

A troca `immutable.List[E]` → `iter.Seq[E]` e a remoção do import de
`functional/immutable` em reducer/projection são **quebra de contrato
público**. Tags `modeling` existem até `modeling/v0.6.0` com o contrato
anterior baseado em List; consumidores dessas tags precisam adaptar chamadas
(fornecer `iter.Seq[E]`, tipicamente via `Events()` ou `list.Values()`).

Rollback após adoção é nova decisão de compatibilidade e release SemVer
major/minor conforme política do módulo. Não se autoriza dual-API
List+Seq para “suavizar” a quebra nesta linha.

## Consequências

- Aplicações evitam `CollectList` no caminho Decision → Replay/Project.
- Qualquer `iter.Seq[E]` (lazy ou materializado via `.Values()`) é entrada
  válida; a ordem de encontro é a da sequência fornecida.
- Concatenação de sequências e multi-pass ficam sob responsabilidade da
  aplicação; as leis de concatenação continuam válidas quando a aplicação
  define `A ++ B` de forma observável.
- `hugeParam` e o restante do lint permanecem ativos: valores de estado ou
  projeção grandes são tratados na modelagem da aplicação, sem
  `Reducer[*S, E]`, `Ref[T]` ou `//nolint` no metamodelo.
- Superfície excluída de ADR-006/007 permanece excluída (Apply, Fallible*,
  registries, persistence, runtime ES/CQRS, etc.).
