---
type: adr
id: adr:workflow:001
context: workflow
title: Workflow as reusable effectful composition
status: Aceita
tags: [workflow, composition, effects, generics, package-boundary]
deciders: [alvadorn]
related: [adr:workflow:002, adr:workflow:003, tdd:workflow:core]
date: 2026-08-10
---
# ADR-001: Workflow como composição reutilizável com efeitos

## Contexto e problema

O código de aplicação mistura repetidamente três momentos não relacionados: adquirir
dependências, receber uma entrada e executar efeitos. Quando os três colapsam em uma
única função, a operação não pode ser reutilizada com entradas diferentes, testada
sem infraestrutura, nem composta com uma operação vizinha sem duplicá-la.

## Drivers da decisão

- reutilização de uma operação composta como valor de primeira classe;
- entradas separadas das dependências;
- execução separada da descrição;
- interoperabilidade com funções Go ordinárias;
- ausência de runtime, contêiner ou registro.

## Decisão

`Workflow[I, O] = func(I) action.Action[O]` é a abstração central do módulo.
As dependências são capturadas na construção do workflow, a entrada chega quando ele
é aplicado, e o contexto chega apenas na execução. A composição é binária e
preserva tipos: `Then(Workflow[A,B], Workflow[B,C]) Workflow[A,C]`.

## Alternativas consideradas

- **Uma interface com método `Execute`.** Rejeitada: força um tipo nominal por
  operação, bloqueia os combinadores de função livre e transforma a injeção de
  capacidades em encanamento de construtores.
- **Um `Then(a, b, c, d)` variádico.** Rejeitado: os genéricos de Go não preservam
  um tipo intermediário que muda ao longo de uma lista variádica, e os únicos
  contornos são `any`, reflection ou asserções de tipo em runtime.
- **Esconder `Action` atrás de `Workflow`.** Rejeitado: as capacidades deixariam de
  poder retornar um efeito componível, o que é o que torna a injeção funcional
  viável.

## Consequências

Um `Workflow` não é handler, controller, service object, job, workflow durável nem
máquina de estados, embora qualquer um desses possa executar um. Workflows compostos
são indistinguíveis dos primitivos, então o consumidor nunca aprende a árvore de
composição. Como um `Workflow` não guarda estado de execução, um valor serve a
qualquer número de entradas, contextos e goroutines.

## Não-objetivos

Persistência de workflow, resume, replay, identificadores de workflow, timers
duráveis e agendamento estão permanentemente fora de escopo; isto é composição
em processo, não um motor de orquestração.
