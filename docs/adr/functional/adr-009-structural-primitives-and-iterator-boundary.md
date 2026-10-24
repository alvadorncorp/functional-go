---
type: adr
id: adr:functional:009
context: functional
title: Tuplas, lazy, iteradores e optics sobre contratos nativos de Go
status: Aceita
tags: [functional, generics, iterators, concurrency, optics, dependencies]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:008, adr:functional:014]
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
- Evitar acoplamento a immutable, state, modeling ou internals.
- Tornar qualquer aresta pública dirigida, mínima e mecanicamente testável.
- Não criar abstrações de efeitos, optics acadêmicas ou iteradores stateful.

## Opções Consideradas

### Opção 1: Omitir iterator e usar somente a stdlib

Preserva a menor superfície, mas repete Map, Filter, FlatMap, limites, Zip,
fold e buscas em consumidores que já expõem `iter.Seq`.

### Opção 2: Criar um tipo Iterator com Next/Value

Centraliza operações, mas perde interoperabilidade direta com range-over-func,
`slices`, `maps` e os iteradores já expostos por immutable.

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

`lazy` é um handle por ponteiro: executa no máximo uma vez, é segura
concorrentemente, memoiza sucesso e panic como `sync.OnceValue`, libera a
função inicializadora armazenada e repete o mesmo payload de panic. A
implementação original usava estado explícito sobre `sync.Once`; a emenda
abaixo a troca por `sync.OnceValue`. `optics` inicia somente com Lens tipada e
composição explícita.

## Consequências

**Positivas:** interoperabilidade nativa, ausência de coleções intermediárias,
uma única nova aresta pública e APIs pequenas com ceilings executáveis.

**Negativas:** iterator mantém wrappers nominais para dois adapters de slices
(agora `Deprecated`, ver emenda); Zip herda overhead fixo de Pull e o consumo
assimétrico documentado; Lazy retém o valor ou panic memoizado durante sua
própria vida.

**Neutras / Rollback:** cada pacote é aditivo. Reverter as fases remove código,
guia e inventário correspondente. A remoção de iterator também remove a única
nova aresta. Prism, Traversal, async iterators, effects e caches continuam fora.

## Emenda parcial (2026-10-23, ADR-014)

Emenda parcial que não substitui esta decisão; o status continua **Aceita**.
Registrada com a decisão [ADR-014](adr-014-v1-surface-hygiene.md).

- **Sinônimos `Deprecated`, sem remoção.** `iterator.FromSlice` e
  `iterator.FromValues` (substituto `slices.Values`), `iterator.Collect`
  (substituto `slices.Collect`) e `tuple.FromValues2` (substituto `Of2`)
  continuam exportados e com a mesma semântica; passam a carregar o marcador
  `Deprecated:`. Os adaptadores de origem seguem como visões vivas do slice do
  chamador. O restante da API de `iterator` e `tuple` não muda.
- **`Lazy` sobre `sync.OnceValue`.** O texto original descrevia estado explícito
  protegido por `sync.Once` que "espelha" `sync.OnceValue`. Agora o `Lazy` usa
  `sync.OnceValue` diretamente. Sucesso, panic comum, execução única sob
  concorrência e liberação do inicializador permanecem. A correção: com a
  implementação manual, um `runtime.Goexit` no inicializador deixava o `Lazy`
  concluído com o valor zero e todo `Get` seguinte o devolvia em silêncio; com
  `sync.OnceValue`, o próprio `Get` e os seguintes entram em panic; sob a
  conversão padrão de `panic(nil)` do Go (módulo principal com `go >= 1.21` e
  `GODEBUG panicnil` sem override) o payload é `*runtime.PanicNilError`. Sob o
  override legado `panicnil=1`, `sync.OnceValue` segue o runtime: nunca devolve
  o valor zero como inicializado, mas o tipo do payload não é garantido. Não se
  afirma mudança no tratamento padrão de `panic(nil)`.
