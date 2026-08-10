---
type: adr
id: adr:functional:008
context: functional
title: Primitivas funcionais e fronteira de dependência de fold
status: Aceita
tags: [functional, generics, dependencies, fold]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:007]
date: 2026-08-09
---
# ADR-008: Primitivas funcionais e fronteira de dependência de fold

## Contexto e Problema

O módulo `functional` precisa de quatro pacotes públicos pequenos para composição
de funções, álgebra de predicados, ordem total e redução de slices. Os três
primeiros podem permanecer folhas da stdlib. `fold`, porém, precisa representar
ausência em `Reduce` e cardinalidade mínima em `ReduceNonEmpty` sem duplicar
tipos já estabilizados em `functional/values`.

ADR-003 e ADR-007 preservam a fronteira sem dependências externas e registram
`values/validated -> values/nonempty` como a única aresta pública então
necessária. A nova API exige decidir se a regra será relaxada globalmente ou se
duas novas arestas serão autorizadas de forma dirigida e testável.

## Fatores de Decisão

- Manter `function`, `predicate` e `order` como folhas da biblioteca padrão.
- Reutilizar `values/optional.Optional[T]` e `values/nonempty.NonEmpty[T]` nas
  assinaturas em vez de criar representações concorrentes.
- Não acoplar as primitivas a `persistent`, `modeling`, internals, `reflect` ou
  `unsafe`.
- Preservar mecanicamente a direção e o conjunto exato de arestas públicas.

## Opções Consideradas

### Opção 1: Manter fold independente e retornar `(T, bool)`

**Prós:**
- Nenhuma nova aresta entre pacotes públicos.

**Contras:**
- Abandona a integração tipada requerida para `Reduce`.
- Duplica no call site a semântica que `values/optional` já possui.

### Opção 2: Relaxar a regra para permitir imports arbitrários entre primitivas

**Prós:**
- Facilita reutilização futura sem alterar o archtest.

**Contras:**
- Cria acoplamento sem necessidade demonstrada.
- Enfraquece a auditabilidade da DAG e permite ciclos futuros.

### Opção 3: Autorizar somente duas arestas de fold para values

**Prós:**
- Reutiliza contratos estabilizados de ausência e não-vazio.
- Mantém todas as demais primitivas como folhas e preserva um allow-list exato.
- Permite `ReduceNonEmpty` sem carregar ausência desnecessária.

**Contras:**
- `fold` passa a carregar dois tipos externos ao próprio pacote em sua API v0.
- O archtest e a documentação da DAG precisam ser mantidos juntos.

## Decisão

**Opção escolhida: autorizar somente duas arestas de fold para values.**

`function`, `predicate` e `order` dependem apenas da biblioteca padrão. `fold`
pode importar exatamente `functional/values/optional` e
`functional/values/nonempty`. Nenhuma primitiva importa outra primitiva.
`TestPublicDependencyDAG`, `TestPrimitiveDependencyDAG` e
`TestPrimitivesNoReflectOrUnsafe` tornam o conjunto executável.

ADR-007 permanece como registro histórico e regra de saída da família values:
somente `values/validated` importa `values/nonempty`; nenhum pacote values passa
a importar `fold`. Esta decisão também não amplia a restrição pré-v1 de
assinaturas de bibliotecas registrada em ADR-003, que permanece específica ao
contrato ali decidido; uma política geral para todas as primitivas exigirá outra
decisão explícita.

## Consequências

**Positivas:**
- `Reduce` representa vazio no tipo e `ReduceNonEmpty` elimina ausência quando a
  entrada já garante cardinalidade.
- A DAG continua pequena, dirigida e mecanicamente verificável.

**Negativas:**
- Mudanças futuras em Optional ou NonEmpty podem exigir uma alteração
  coordenada da API pública de fold enquanto o módulo permanece v0.

**Neutras / Ações de acompanhamento:**
- Rollback remove o pacote fold, suas duas arestas allow-listed e esta ADR sem
  afetar os três pacotes-folha.
- Traverse, Sequence, adapters de persistent e abstrações algébricas continuam
  fora do escopo.
