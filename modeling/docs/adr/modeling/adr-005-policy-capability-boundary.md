---
type: adr
id: adr:modeling:005
context: modeling
title: Capability pública e aditiva de Policy
status: Aceita
tags: [modeling, policy, generics, optional, predicate, architecture]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:002, adr:modeling:003, tdd:modeling:v0-6-policy]
date: 2026-08-10
---
# ADR-005: Capability pública e aditiva de Policy

## Contexto e problema

O baseline entregue de modeling está em v0.3 e expõe primitives na raiz. Uma
Policy responde qual resultado ou estratégia deriva de um contexto; ela não
valida, não evolui estado e não executa efeitos. O pedido exige uma capability
própria `modeling/policy`, sem aliases na raiz, e seleção condicional por
Predicate e Optional. Isso contradiz deliberadamente o limite root-only e a
proibição geral de tipos functional em APIs públicas.

O ID 004 permanece reservado ao predecessor planejado Reducer/Replay; Policy usa
005 sem reescrever ou antecipar Change/Reducer/Replay.

## Direcionadores

- manter funções puras, tipos de primeira classe e um único input genérico;
- dar valor próprio à capability por transformação e seleção condicional;
- representar não aplicabilidade como Optional, nunca error;
- preservar todas as APIs v0.1–v0.3 e impedir runtime/dispatch dinâmico;
- limitar a exceção functional aos dois tipos realmente necessários.

## Opções consideradas

### Reexportar Policy na raiz

Rejeitada: amplia a raiz, apaga a capability como unidade e contradiz o pedido.

### Criar um tipo nominal Conditional

Rejeitada por enquanto: funções `I → Optional[O]` bastam e evitam cristalizar
uma segunda abstração sem uso independente demonstrado.

### Usar error para não aplicabilidade

Rejeitada: mistura ausência de seleção com falha de domínio.

### Novo package com exceção funcional estreita (escolhida)

Escolhida: `modeling/policy` expõe exatamente Policy, Constant, Map, Contramap,
When, FirstApplicable e Fallback. Somente suas assinaturas condicionais podem
usar `predicate.Predicate` e `optional.Optional`.

## Decisão

A dependência é application → modeling/policy → os packages públicos
`functional/primitives/predicate` e `functional/values/optional`. A raiz
`modeling`, siblings e `functional` não dependem de policy.

FirstApplicable exige primeiro operando, avalia na ordem declarada e para no
primeiro Some. Fallback só avalia sua policy total em None. Combinadores rejeitam
operandos nil na construção com panics privados determinísticos. Apply,
Conditional nominal, multi-aridade, métodos, aliases e runtime ficam excluídos.

## Compatibilidade e rollback

A mudança é aditiva e opt-in. Nenhum nome, assinatura, comportamento ou import
path existente muda; `functional v1.1.0` preserva os contratos necessários e é
o release fixado.
Antes de adoção, rollback é reverter os commits das fases. Depois de publicação
e adoção, remoção ou mudança das sete assinaturas é breaking e exige nova ADR
e release; este trabalho não publica versão.

## Consequências

Consumidores ganham composição de policies sem engine. Os gates passam a ser
package-aware: preservam a proibição geral e permitem apenas Predicate/Optional
em policy. O plano pendente de Reducer/Replay compartilha esses gates e não pode
ser executado em paralelo; o segundo plano executado deve ser aprofundado.
