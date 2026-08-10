---
type: adr
id: adr:functional:009
context: functional
title: Tuplas, lazy, iteradores e optics sobre contratos nativos de Go
status: Aceita
tags: [functional, generics, iterators, concurrency, optics, dependencies]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:008]
date: 2026-08-09
---
# ADR-009: Tuplas, lazy, iteradores e optics sobre contratos nativos de Go

## Contexto e Problema

O módulo `functional` precisa eliminar boilerplate recorrente para transportar
valores heterogêneos, adiar computações memoizadas, compor sequências lazy e
atualizar valores aninhados de forma imutável. Go 1.26.4 já fornece
`iter.Seq`, `iter.Seq2`, `iter.Pull`, `slices.Values` e `slices.Collect`, mas
não fornece a camada de Map/Filter/FlatMap/limites/Zip/fold/busca requerida.

ADR-003 exige zero dependências externas e uma allow-list exata entre pacotes
públicos. ADR-008 mantém function, predicate e order como folhas e autoriza
somente as duas arestas de fold para values. Os novos pacotes precisam preservar
essa disciplina sem criar um Iterator concorrente com a stdlib.

## Fatores de Decisão

- Reutilizar contratos nativos de Go e range-over-function.
- Manter tipos pequenos, genéricos, previsíveis e sem reflect/unsafe.
- Evitar acoplamento a persistent, state, modeling ou internals.
- Tornar qualquer aresta pública dirigida, mínima e mecanicamente testável.
- Não criar abstrações de efeitos, optics acadêmicas ou iteradores stateful.

## Opções Consideradas

### Opção 1: Omitir iterator e usar somente a stdlib

Preserva a menor superfície, mas repete Map, Filter, FlatMap, limites, Zip,
fold e buscas em consumidores que já expõem `iter.Seq`.

### Opção 2: Criar um tipo Iterator com Next/Value

Centraliza operações, mas perde interoperabilidade direta com range-over-func,
`slices`, `maps` e os iteradores já expostos por persistent.

### Opção 3: Pacotes pequenos e iterator como adapters de iter.Seq

Mantém tuple, lazy e optics como folhas da stdlib; iterator recebe e devolve
`iter.Seq` e importa somente tuple para representar Zip e Enumerate como um
único valor composável.

## Decisão

**Escolhida a opção 3.** Criar `tuple`, `lazy`, `iterator` e `optics` no módulo
existente. Não criar facade no pacote raiz. `tuple`, `lazy` e `optics` dependem
somente da stdlib. `iterator` depende da stdlib e exatamente de `tuple`.

`iterator` não define tipo Iterator. FromSlice e Collect delegam a
`slices.Values` e `slices.Collect`; as demais operações preservam laziness e
early termination. Zip usa `iter.Pull`, encerra ambos os pulls e termina no
menor input. A ordem de pull é esquerda e depois direita; se a direita terminar
primeiro, no máximo um valor esquerdo sem par pode ter sido consumido.

`lazy` é um handle por ponteiro com estado explícito protegido por `sync.Once`:
executa no máximo uma vez, é segura concorrentemente, espelha a memoização de
sucesso/panic de `sync.OnceValue`, libera a função inicializadora armazenada e
repete o mesmo payload de panic. `optics` inicia somente com Lens tipada e
composição explícita.

## Consequências

**Positivas:** interoperabilidade nativa, ausência de coleções intermediárias,
uma única nova aresta pública e APIs pequenas com ceilings executáveis.

**Negativas:** iterator mantém wrappers nominais para dois adapters de slices;
Zip herda overhead fixo de Pull e o consumo assimétrico documentado; Lazy retém
o valor ou panic memoizado durante sua própria vida.

**Neutras / Rollback:** cada pacote é aditivo. Reverter as fases remove código,
guia e inventário correspondente. A remoção de iterator também remove a única
nova aresta. Prism, Traversal, async iterators, effects e caches continuam fora.
