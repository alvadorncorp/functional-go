---
type: adr
id: adr:workflow:004
context: workflow
title: Explicit transaction scopes and scoped capabilities
status: Aceita
tags: [workflow, transaction, capabilities, context, boundary]
deciders: [alvadorn]
related: [adr:workflow:003, adr:workflow:006, adr:workflow:008, tdd:workflow:transaction]
date: 2026-08-10
---
# ADR-004: Escopos de transação explícitos e capacidades escopadas

## Contexto e problema

Um workflow precisa escrever várias coisas atomicamente sem saber qual tecnologia
fornece atomicidade, e sem que uma escrita escapulida saia da transação. O atalho
comum — esconder o handle da transação em `context.Context` — torna essa fuga
invisível e intipável.

## Drivers da decisão

- o handle concreto nunca chegar a uma assinatura de workflow;
- escritas que não possam rodar acidentalmente fora da fronteira;
- plena segurança de tipos para o resultado, apesar dos limites dos genéricos de Go;
- uma alegação honesta de atomicidade.

## Decisão

`Boundary[S] = func(context.Context, func(context.Context, S) error) error` é a
região fornecida pelo adaptador, e a função livre
`Within[S, T](boundary, func(S) action.Action[T]) action.Action[T]` executa uma
computação tipada dentro dela. O escopo `S` é definido pela aplicação e carrega
capacidades, nunca um handle. `context.Context` carrega ciclo de vida —
cancelamento e prazos — e nunca uma transação.

## Alternativas consideradas

- **Um método `Within` em um tipo de transação.** Rejeitado: Go proíbe um método
  introduzir o parâmetro de tipo extra `T`, e todo contorno custa `any`,
  reflection ou asserções de tipo. Segurança de tipos vence forma fluente.
- **`Action[T] -> Action[T]` como forma da fronteira.** Rejeitado: a computação
  interna não teria como alcançar capacidades ligadas à transação.
- **Um registro genérico de escopos indexado por nome.** Rejeitado: `map[string]any`
  apaga exatamente os tipos que esta decisão existe para preservar.

## Consequências

O estreitamento de capacidades torna-se expressável: a aplicação pode oferecer
apenas capacidade de leitura fora da fronteira e apenas capacidades de escrita
dentro dela, de modo que uma escrita fora do lugar falha ao compilar. `Boundary`
em si não promete ACID — o adaptador documenta sua garantia. Chamadas aninhadas
de `Within` não têm significado especial; savepoints e transações aninhadas
pertencem ao adaptador.

## Não-objetivos

Sem savepoints, sem `UnitOfWork` (que implica identity maps, change tracking e
ciclos de vida de entidade que este desenho não tem) e sem transação ambiente ou
padrão.
