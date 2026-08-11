---
type: adr
id: adr:workflow:010
context: workflow
title: Plan as explicit non-executable workflow description
status: Aceita
tags: [workflow, plan, description, tooling, immutability]
deciders: [alvadorn]
related: [adr:workflow:001, adr:workflow:002, adr:workflow:009]
date: 2026-08-11
---
# ADR-010: Plan como descrição explícita e não executável de workflow

## Contexto e problema

`Workflow[I,O]` constrói uma `Action[O]`, e `Action` continua sendo uma closure
opaca executada somente por `Run`. Documentação, visualização e testes
arquiteturais precisam descrever a estrutura conceitual de uma operação sem
transformar essas closures em AST, runtime ou DSL executável.

## Decisão

Criamos `workflow/plan` como pacote stdlib-only. `Plan` é dado imutável,
declarado explicitamente pela aplicação e inspecionado por `Kind`, `Name` e
`Children`. A superfície inicial contém somente Empty, Step, Sequence e
Transaction.

O valor zero é Empty. `Step` preserva o nome não vazio. `Sequence()` é Empty,
`Sequence(p)` é `p`, e sequências maiores preservam ordem e aninhamento.
`Transaction` preserva sempre o nó conceitual, inclusive com zero ou um filho.
Construtores e `Children` fazem cópias defensivas.

Plan não recebe contexto, não executa `Action`, não chama `Run`, não abre
transação, não aplica resilience, não inicia goroutine e não realiza I/O.

## Consequências

- Plan pode existir sem Workflow, e Workflow pode existir sem Plan.
- A associação entre ambos permanece opt-in e pertence à aplicação.
- Plan não conhece `action`, `workflow` root, transport ou infrastructure.
- Tooling pode inspecionar uma árvore estável sem observar histórico de execução.

## Alternativas rejeitadas

- **Inferir Plan de closures.** Exigiria reflection, runtime metadata ou unsafe.
- **Compilar Plan para Action.** Criaria uma segunda linguagem executável.
- **Guardar Plan dentro de Workflow.** Quebraria o contrato funcional existente.
- **Adicionar Visitor, Resource, metadata ou JSON agora.** Congelaria superfície
  sem consumo demonstrado.

## Não-objetivos

Plan não é Workflow, Action, `modeling.Specification`, execution history,
runtime AST, scheduler, registry ou durable workflow definition.
