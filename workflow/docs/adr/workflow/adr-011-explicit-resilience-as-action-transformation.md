---
type: adr
id: adr:workflow:011
context: workflow
title: Explicit resilience as Action transformation
status: Aceita
tags: [workflow, action, resilience, retry, timeout]
deciders: [alvadorn]
related: [adr:workflow:002, adr:workflow:009, adr:workflow:010]
date: 2026-08-11
---
# ADR-011: Resiliência explícita como transformação de Action

## Contexto e problema

ADR-002 define Action como computação adiada e ligada ao contexto recebido por
Run. ADR-009 excluiu retry e timeout da primeira álgebra para impedir que
composição avançada virasse engine. Aplicações, porém, precisam aplicar essas
duas políticas no adapter sem escondê-las no domínio, no Workflow ou na Action
de origem.

## Decisão

Criamos `workflow/resilience`, dependente somente de `workflow/action` e da
stdlib. O pacote exporta `Backoff`, `RetryPolicy`, `Retry` e `Timeout`.

`Retry` valida política, predicado e Action durante a construção, mas só chama a
Action em Run. `Attempts` inclui a primeira tentativa. `When` é consultado
apenas depois de uma falha para a qual ainda existe nova tentativa. `Backoff`
recebe o número 1-based da tentativa que acabou de falhar; atraso zero é
imediato, atraso negativo é defeito de programação. Esperas observam
`ctx.Done()`. O primeiro sucesso encerra o loop; falha devolve o zero de T e a
identidade exata do erro da origem ou do contexto. Nenhum panic é recuperado.

`Timeout` valida duração positiva e Action não-nula na construção. Em Run cria
um contexto filho com `context.WithTimeout`, chama a Action sincronamente e
devolve seu resultado sem embrulho. É cancelamento cooperativo, não preempção:
uma Action que ignora contexto não é interrompida por goroutine escondida.

A ordem de composição é pública: `Timeout(Retry(source))` dá um orçamento total;
`Retry(Timeout(source))` cria um orçamento por tentativa. Retry só é correto
quando a operação é idempotente, deduplicada ou quando seus efeitos parciais
foram modelados explicitamente.

## Consequências

- A aplicação escolhe classe de erro, número de tentativas, backoff e orçamento.
- Resilience não conhece HTTP, broker, faults, transação, domínio ou Workflow.
- `time` e `context.WithTimeout` recebem exceção arquitetural somente no pacote.
- Não há goroutine, recover, jitter implícito, telemetria ou estado compartilhado.

## Alteração de ADR-009

Fica superada somente a exclusão absoluta de retry/timeout. Permanecem válidas
as rejeições a engine, runtime, execução em background, paralelismo implícito e
orquestração durável.

## Alternativas rejeitadas

- Embutir política em Action ou Workflow.
- Tornar Timeout preemptivo com goroutine e canal.
- Classificar erros no pacote resilience.
- Exportar combinadores de jitter, circuit breaker, bulkhead ou Policy genérica.

## Não-objetivos

Resilience não promete idempotência, entrega exactly-once, deadline remoto,
rollback, durabilidade, scheduling ou interrupção forçada.
