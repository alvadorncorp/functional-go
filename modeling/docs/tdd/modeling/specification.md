---
type: tdd
id: tdd:modeling:specification
context: modeling
title: Modeling Specification
status: Aceito
tags: [modeling, specification, classification, laws, testing]
related: [adr:modeling:008, adr:modeling:009, tdd:modeling:policy, tdd:modeling:rules-requirements]
date: 2026-08-11
---
# TDD: Modeling Specification

## Escopo técnico

Criar o leaf package público `modeling/specification`, independente da raiz,
de Policy e dos demais siblings. O package é metamodelo puro, sem efeito,
runtime, infraestrutura ou tipo de domínio concreto.

## Contrato público

```go
type Specification[T any] func(T) bool
func All[T any](specifications ...Specification[T]) Specification[T]
func Any[T any](specifications ...Specification[T]) Specification[T]
func Not[T any](spec Specification[T]) Specification[T]
func Contramap[A, B any](spec Specification[B], mapper func(A) B) Specification[A]
```

Essa é a superfície inteira. `Of` e `Xor` foram avaliados e rejeitados;
Filter, Validate, Explain, Reason, Error, Message e bridges ficam excluídos.

## Semântica

Specification classifica um valor com vocabulário de domínio. All e Any
avaliam left-to-right com short-circuit; `All()` retorna true e `Any()` retorna
false. Not nega. Contramap adapta um input antes de uma única avaliação.

A zero Specification é função nil e panica naturalmente quando invocada.
Combinadores rejeitam operandos nil eager com
`modeling/specification: nil specification`; Contramap rejeita mapper nil com
`modeling/specification: nil mapper`. Nil nunca significa true, false ou skip.

## Dependências e arquitetura

Os production files não possuem imports. Specification não importa Predicate,
raiz, Policy, Change, Reducer ou Projection. Archtests pinam o oitavo package,
zero imports, cinco exports, ausência de aliases na raiz e proibição das arestas
entre siblings.

## Testes e leis

Testes black-box fixam identidade nominal, atribuição de função ordinária,
zero/nil, ordem, short-circuit e Contramap. Leis observacionais cobrem dupla
negação, formas unárias, identidades vazias e true/false. Fuzzing da standard
library valida as mesmas propriedades sobre inputs arbitrários sem comparar
funções diretamente nem adicionar dependência.

## Verificação

Executar testes focados, archtests, módulo, módulo standalone, race, vet,
tidy-diff, fuzz bounded, exemplo externo, workspace tests e lint uncapped.

## Compatibilidade e exclusões

A capability é aditiva. Nenhuma API existente muda. Não há go.mod, go.work,
benchmark, migration, config, release, reflection, error, runner, engine,
registry, dispatcher, persistence, transporte ou efeito.
