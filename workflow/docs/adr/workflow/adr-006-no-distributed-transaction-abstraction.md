---
type: adr
id: adr:workflow:006
context: workflow
title: No distributed transaction abstraction
status: Aceita
tags: [workflow, transaction, consistency, bounded-context, outbox]
deciders: [alvadorn]
related: [adr:workflow:004, adr:workflow:008, tdd:workflow:transaction]
date: 2026-08-10
---
# ADR-006: Sem abstração de transação distribuída

## Contexto e problema

Uma operação frequentemente atravessa dois bounded contexts. É tentador oferecer
uma API que aceite várias fronteiras e pareça torná-las atômicas juntas.

## Drivers da decisão

- uma API que nunca prometa uma garantia que não possa entregar;
- suporte a ambas as topologias de implantação sem confundi-las.

## Decisão

`Boundary[S]` opera sobre exatamente uma região transacional concreta. Nenhuma API
aceita múltiplas fronteiras, e os nomes `DistributedTransaction`,
`GlobalTransaction`, `TwoPhaseCommit` e `Saga` são nomes exportados proibidos.

Dois bounded contexts que compartilham um resource manager podem legitimamente
confirmar em uma transação ACID: bounded context é fronteira semântica e
transação é fronteira física, e nada exige que coincidam. Dois contextos em
resource managers independentes obtêm atomicidade local mais outbox mais
consistência eventual.

## Alternativas consideradas

- **Um combinador que compõe duas fronteiras.** Rejeitado: sequenciar `COMMIT A`
  e depois `COMMIT B` não é atômico, e uma falha após o primeiro deixa o sistema
  parcialmente confirmado. Apresentar isso como transação seria incorreto.
- **XA / two-phase commit.** Rejeitado para o núcleo: exige suporte do resource
  manager, um coordenador e semântica de recuperação muito além de uma biblioteca
  de composição em processo.

## Consequências

O exemplo cross-context embarca ambas as topologias e as rotula explicitamente.
Na distribuída, uma falha do Ledger deixa Banking confirmado e a mensagem
pendente, e o teste afirma exatamente isso em vez de fingir que não aconteceu.
A entrega é at-least-once, então a idempotência do consumidor é responsabilidade
da aplicação.

## Não-objetivos

Nenhum framework de compensação. Um futuro `workflow/saga` permaneceria
semanticamente distinto de `workflow/transaction`: rollback é o resource manager
desfazendo trabalho não confirmado; compensação é uma nova operação revertendo
trabalho já confirmado.
