---
type: adr
id: adr:modeling:008
context: modeling
title: Capability nominal e independente de Specification
status: Aceita
tags: [modeling, specification, classification, packages, architecture]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:005, adr:modeling:009, tdd:modeling:specification]
date: 2026-08-11
---
# ADR-008: Capability nominal e independente de Specification

> **Correção de redação:** a decisão original descrevia o package como
> "zero-import". A implementação entregue importa somente o stdlib `slices`
> (`logic.go`); a decisão — sem `functional`, raiz, `policy` nem siblings —
> permanece em vigor e o texto abaixo foi alinhado a esse fato.

## Contexto e problema

O módulo distingue condição genérica, classificação de domínio, validação
explicável, pré-condição operacional, invariante e seleção de estratégia.
`functional.Predicate[T]` e `Specification[T]` compartilham a representação
`func(T) bool`, mas não a responsabilidade semântica. A nova capability deve
nomear a classificação de domínio sem ampliar a raiz nem criar engine.

## Direcionadores

- preservar `application -> modeling -> functional` e impedir a direção inversa;
- manter capabilities irmãs independentes e opt-in;
- comunicar intenção por identidade nominal própria;
- oferecer somente álgebra comprovada por exemplos de domínio;
- manter nil, ordem e short-circuit previsíveis antes do congelamento da API.

## Alternativas consideradas

### Reexportar Specification na raiz

Rejeitada: ampliaria a fachada estável e apagaria a fronteira da capability.

### Alias público de Predicate

Rejeitada: alias perderia a identidade nominal e confundiria álgebra genérica
com vocabulário de domínio.

### Adaptar internamente a álgebra de Predicate

Rejeitada: `All` e `Any` exigiriam conversão de slices, os checks eager de nil
continuariam locais e os loops diretos são menores. O acoplamento não compra
semântica nem implementação relevante.

### Leaf package direto, nominal e somente stdlib `slices` (escolhida)

Escolhida: `modeling/specification` expõe exatamente `Specification`, `All`,
`Any`, `Not` e `Contramap` e implementa a álgebra sobre funções Go.

## Decisão

`Specification[T]` é `func(T) bool` nominal e responde se um valor pertence a
uma categoria lógica do domínio. Não valida, não retorna error, não explica
falha, não autoriza operação e não seleciona estratégia.

`All` e `Any` avaliam da esquerda para a direita e fazem short-circuit.
`All()` é a identidade true; `Any()` é a identidade false. `Not` nega e
`Contramap` adapta o input. Combinadores rejeitam Specification ou mapper nil
na construção com panics determinísticos. Invocar a zero Specification mantém
o comportamento natural de função nil de Go.

O package não importa `functional`, raiz `modeling`, `policy` nem siblings; seu
único import de produção é o stdlib `slices` (`logic.go`), e `transform.go` e
`specification.go` não têm imports.
`Of` fica fora porque atribuição/conversão direta é suficiente. `Xor` fica fora
porque nenhum caso de domínio o justificou. Filter, error, mensagem, reason,
validate, registry, reflection, runtime e bridges também ficam excluídos.

## Compatibilidade e rollback

A mudança é aditiva. Antes de publicação, rollback é reverter os commits da
capability. Depois de publicação, remover o package ou alterar os cinco
contratos, identidades vazias, ordem, short-circuit ou nil semantics é breaking
e exige nova decisão e tratamento SemVer. Este trabalho não publica release.

## Consequências

Aplicações ganham classificação nominal e composável mantendo Policy,
Requirement, Rule e Invariant semanticamente distintos. O custo é uma pequena
duplicação intencional da álgebra booleana genérica em troca de uma fronteira
de dependências mais forte (leaf somente stdlib `slices`) e de contratos de nil
próprios.
