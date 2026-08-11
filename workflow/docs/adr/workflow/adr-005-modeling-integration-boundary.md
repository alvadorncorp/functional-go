---
type: adr
id: adr:workflow:005
context: workflow
title: Modeling integration boundary
status: Aceita
tags: [workflow, modeling, domain, boundary, purity]
deciders: [alvadorn]
related: [adr:workflow:001, adr:workflow:007, tdd:workflow:modeling-integration]
date: 2026-08-10
---
# ADR-005: Fronteira de integração com modeling

## Contexto e problema

`modeling` é dono da semântica de domínio pura — `Requirement`, `Transition`,
`Invariant`, `Decision`, `Policy`. `workflow` é dono da coordenação com efeitos.
A questão é qual módulo depende de qual, e se `workflow` deve embarcar ajudantes
para `Decision`.

## Drivers da decisão

- `modeling` permanecer puro e alheio à execução;
- nenhuma duplicação do metamodelo de domínio dentro de `workflow`;
- evidência antes de abstração.

## Decisão

A aresta de dependência `workflow -> modeling` não é criada. A aplicação é a
camada de composição que importa ambos, o que os torna módulos irmãos e não uma
pilha. `modeling -> workflow` é proibido e controlado por gate. `workflow` não
define `Rule`, `Requirement`, `Transition`, `Invariant`, `Policy` nem `Decision`.

## Alternativas consideradas

- **Uma ponte `workflow/integration/modeling` nesta entrega.** Adiada, não
  rejeitada: os exemplos de saque e cross-context mostraram que o boilerplate
  carregar-decidir-confirmar difere o suficiente entre domínios para que um
  `CommitDecision[S,E,O]` compartilhado codificasse um estilo arquitetural como se
  fosse geral. A extração espera pelo menos três domínios concordando na forma.
- **`workflow` importar `modeling` por conveniência.** Rejeitado: acoplaria as
  linhas de versão e não compraria nada que a aplicação não possa expressar.

## Consequências

A forma canônica é carga com efeito, decisão pura, confirmação com efeito, com o
workflow chamando uma função de domínio ordinária como
`domain.Withdraw(account, cmd) (modeling.Decision[Account, Event], error)`.
Limiares de negócio vivem no domínio; um workflow que ramifica o faz sobre um
valor que o domínio produziu.

## Não-objetivos

Nenhuma premissa de event sourcing, nenhuma abstração de aggregate e nenhuma
exigência de que um `Decision` seja persistido.
