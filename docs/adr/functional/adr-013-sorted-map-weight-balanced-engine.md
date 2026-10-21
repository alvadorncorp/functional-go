---
type: adr
id: adr:functional:013
context: functional
title: SortedMap sobre um motor de árvore balanceada por peso (wbtree)
status: Aceita
tags: [functional, immutable, sorted-map, wbtree, engine, ownership]
deciders: [alvadorn]
related: [adr:functional:004, adr:functional:005]
date: 2026-10-21
---
# ADR-013: SortedMap sobre um motor de árvore balanceada por peso (wbtree)

## Contexto e Problema

`immutable.Map` e `immutable.Set` usam a HAMT do ADR-005: a ordem de
iteração não é especificada e muda entre linhagens, porque cada raiz tem a
sua própria `maphash.Seed`. Não existe coleção imutável com ordem de chave
definida pelo usuário.

O item 3b (um mapa imutável ordenado) foi aprovado pelo usuário. Sendo
honesto sobre a demanda: hoje nenhum consumidor de `Map` precisa de ordem, e
o caso citado na proposta (emitir os elementos de um `Set` em ordem
determinística) já é atendido por `slices.Sorted(s.Values())`, ou
`slices.SortedFunc` quando `T` não é `cmp.Ordered`. Esse continua sendo o
caminho mais leve para uma emissão pontual. `SortedMap` existe para quem
mantém o mapa ordenado ao longo de várias versões, sem reordenar a cada
leitura.

## Fatores de Decisão

- Atualização e leitura em `O(log n)`, com path-copy e compartilhamento
  estrutural, como `Map`.
- Builders com edição transitória sob `*ownership.Owner` e Freeze `O(1)`
  (ADR-005), sem caminhada na árvore para publicar.
- Núcleo fino intacto (ADR-004): o motor só importa `ownership` e `cow`.
- Zero dependências fora da stdlib, sem `reflect` nem `unsafe`.
- A menor árvore balanceada cujo balanceamento tem prova publicada.

## Opções Consideradas

### Opção 1: B-tree persistente

Rejeitada por ora: melhor localidade de cache, mas split/merge de nós largos
com dono transitório é bem mais código, e cada path-copy copia nós inteiros.
Fica como troca possível do motor, sem mudar a API, só se os benchmarks de
`SortedMap` mostrarem `Get`/`Set` mais de cerca de 3x mais lentos que `Map`
com 10k chaves.

### Opção 2: Red-black ou AVL persistente

Rejeitada: a remoção persistente em red-black tem muitos casos, e nenhuma
das duas guarda o tamanho da subárvore. A árvore por peso guarda esse
tamanho de graça, o que dá `Len` em `O(1)` e deixa rank, split e join
possíveis no futuro.

### Opção 3: Só ordenar na leitura

Manter `Map` e ordenar em cada iteração. Custa `O(n log n)` por leitura e
não cria tipo novo. Rejeitada como única resposta ao item 3b, mas continua
sendo a receita recomendada para emissões pontuais (`slices.Sorted`).

### Opção 4: Árvore balanceada por peso com os parâmetros de Straka (escolhida)

## Decisão

Adicionar `immutable.SortedMap[K, V]` e `immutable.SortedMapBuilder[K, V]`
sobre um novo motor interno `internal/engine/wbtree`.

### Motor `wbtree`

Árvore binária de busca persistente balanceada por peso (Adams), com os
parâmetros provados por Straka ("Adams' Trees Revisited", 2011) sobre o
peso `size+1`, como constantes nomeadas:

```
delta = 3
gamma = 2

isBalanced(a, b) := delta*(size(a)+1) >= size(b)+1
isSingle(a, b)   := size(a)+1 < gamma*(size(b)+1)
```

Todo nó satisfaz `isBalanced(left, right)` e `isBalanced(right, left)`.
Depois de uma inserção ou remoção, cada nível do caminho é reparado por uma
rotação simples (quando `isSingle(interno, externo)` vale para os netos do
lado pesado) ou dupla. As rotações olham só tamanhos, nunca o comparador.

- Operações puras fazem path-copy e publicam nós com `owner == nil`.
- `TransientSet`/`TransientDelete` recebem `*ownership.Owner`, editam em
  lugar os nós desse dono ativo e clonam os demais via `cow.EnsureEditable`.
  O motor não importa `transient`; publicar é `Invalidate` em `O(1)`.
- Toda comparação de um caminho acontece na descida, antes de qualquer nó
  ser editado na volta. Um comparador que entra em pânico não deixa nós do
  dono pela metade, e o builder mantém o conteúdo anterior.
- Remover uma chave ausente devolve o mesmo ponteiro de nó.
- Imports permitidos: `ownership` e `cow`. O archtest de `functional`
  registra o motor e essa allow-list.

### Contrato do comparador

`cmp func(a, b K) int` devolve negativo, zero ou positivo e precisa ser uma
ordem fraca estrita: determinístico, transitivo e estável enquanto o mapa
existir. Chaves com `cmp(a, b) == 0` são a mesma chave. O comparador é fixado
na construção e herdado por toda versão derivada, inclusive uma versão
esvaziada por `Delete`. `EmptySortedMap(nil)` e `CollectSortedMap(nil, seq)`
entram em pânico com `immutable: nil comparator`. Um pânico do comparador se
propaga ao chamador. A iteração (`All`, `Keys`, `Values`) é crescente pelo
comparador; a ordem de `Map` e `Set` continua não especificada.

### Exceção do valor zero

`K any` não tem ordem padrão, então um `SortedMap` zero não tem comparador.
Ele lê como vazio (`Len`, `Get`, `Has`, iteração) e `Delete` nele é no-op,
mas `Set` e `Builder()` entram em pânico com exatamente:

```
immutable: zero SortedMap has no comparator; construct with EmptySortedMap or CollectSortedMap
```

É uma exceção deliberada à paridade com `Map` e `Set`, cujo valor zero
aceita escrita. Um getter seguro para valor zero precisa devolver
`EmptySortedMap(cmp)`. Mensagens `wbtree:` nunca chegam ao usuário:
`immutable` valida antes de chamar o motor.

### Regra de substituição da chave

`Set` substitui a chave guardada e o valor quando já existe uma chave
equivalente, em todos os caminhos (puro, builder e `CollectSortedMap`). A
última chave e o último valor vencem, e isso é observável por `Keys()` com
um comparador que ignora maiúsculas, por exemplo.

### Escopo da API

Construtores `EmptySortedMap(cmp)` e `CollectSortedMap(cmp, seq)`; métodos
`Len`, `IsEmpty`, `Get`, `Has`, `Set`, `Delete`, `All`, `Keys`, `Values`,
`Equal`, `EqualFunc` e `Builder`; o builder tem `Set`, `Delete`, `Len`,
`Get`, `Has` e `Build`, que é `O(1)`. Ficam fora: `First`, `Last`,
`Floor`, `Ceiling`, consultas por intervalo e iteração reversa;
`SortedSet`; `SortedMap.Update`; e as operações rank, split e join do
motor. São aditivas e não têm consumidor hoje.

## Consequências

**Positivas:** coleção ordenada com path-copy em `O(log n)`, compartilhamento
estrutural e Freeze `O(1)`, sem tocar o núcleo nem os motores existentes. A
regra da chave fica única e testada. O tamanho por nó deixa rank e split
possíveis sem trocar de motor.

**Negativas:** um terceiro motor para manter. O valor zero de `SortedMap`
quebra a paridade com `Map` e `Set` na escrita. Comparadores inconsistentes
corrompem a ordem sem diagnóstico. Uma árvore binária tem pior localidade de
cache que HAMT ou B-tree.

**Verificação:** checador de invariantes com os predicados exatos após cada
operação, teste sequencial crescente e decrescente com cerca de 10k chaves,
`FuzzOps` contra um modelo, testes de transitório e compartilhamento, limite
de alocações de um `Set`, testes de aceitação em `immutable_test`, archtest
de dependências do motor, `go test ./...` e lint.
