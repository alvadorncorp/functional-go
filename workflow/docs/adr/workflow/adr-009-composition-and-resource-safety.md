---
type: adr
id: adr:workflow:009
context: workflow
title: Advanced composition and normal-return resource safety
status: Aceita
tags: [workflow, action, composition, resources, errors, context]
deciders: [alvadorn]
related: [adr:workflow:002, adr:workflow:004, adr:workflow:007, tdd:workflow:action, tdd:workflow:composition]
date: 2026-08-11
---
# ADR-009: Composição avançada e resource safety em retornos normais

## Contexto e problema

`Workflow` e `Action` já descrevem efeitos adiados, reexecutáveis e ligados ao
`context.Context`, mas a composição ainda exige wrappers manuais para adaptar
inputs, combinar computações independentes, observar sucesso, transformar erro
em dado e delimitar o ciclo de vida de um recurso.

## Drivers da decisão

- preservar `Run` como única passagem da descrição para a execução;
- ampliar a expressão sem criar engine, runtime, scheduler ou container;
- manter o núcleo stdlib-only e o DAG `workflow -> action <- transaction`;
- tornar ordem, curto-circuito, finalização e dupla falha observáveis por testes;
- manter infraestrutura e política de cleanup nos adapters.

## Decisão

O root recebe `Contramap` para adaptar o tipo de entrada e `Provide` para fixar
uma entrada e obter uma `Action`. Aplicar um workflow continua sem executar o
efeito; somente `Run` executa a Action resultante.

`action` recebe `Pair`, `Zip`, `Both`, `Tap`, `Ensure`, `AttemptResult`,
`Attempt`, `Bracket` e `AcquireUseRelease`. `Zip` e `Both` sempre executam
primeiro a Action da esquerda e depois a da direita. Independência significa
apenas que o segundo valor não depende do primeiro; não autoriza paralelismo.

`Tap` observa somente sucesso. `Ensure` executa um finalizer depois de todo
retorno normal `(T, error)`. `Attempt` transforma somente o canal normal de erro
em dado. `Pair` e `AttemptResult` permanecem carriers com campos, sem métodos.

`Bracket` é a implementação canônica de acquire/use/release;
`AcquireUseRelease` delega a ela. Release executa se e somente se acquire
retornar sucesso, inclusive depois de um erro normal de use. Erros únicos
mantêm identidade; dupla falha usa `errors.Join(primaryErr, finalizerErr)`.

Nenhum combinador chama `recover`, cria goroutine, deriva contexto ou aplica
`context.WithoutCancel`. Panics continuam indicando erro de programação e não
há garantia de finalização/release no caminho de panic. O mesmo contexto dado a
`Run`, mesmo cancelado, chega ao finalizer/release; shielding pertence ao
adapter.

`Bracket` não substitui `transaction.Within`: resource lifecycle não promete
begin, commit, rollback ou atomicidade.

## Alternativas consideradas

- **Executar `Zip`/`Both` em paralelo.** Rejeitado: exigiria runtime,
  sincronização e política de cancelamento fora do escopo.
- **Usar `defer` + `recover` para cleanup em panic.** Rejeitado: criaria um
  terceiro canal de controle diferente do restante de `Action`.
- **Blindar release com `context.WithoutCancel`.** Rejeitado: é política do
  recurso concreto.
- **Importar tuples/results de `functional`.** Rejeitado: quebraria o fecho
  stdlib-only e duplicaria responsabilidade.
- **Fundir `Bracket` e `Within`.** Rejeitado: lifecycle e transação possuem
  contratos diferentes.

## Consequências

A superfície pública cresce de forma aditiva e precisa ser exercitada por
exemplos compilados. O pacote `action` passa a importar `errors` além de
`context`, ainda dentro da biblioteca padrão. Arquitetura e testes comportamentais
fixam a ordem sequencial, carriers mínimos, ausência de goroutines/recovery,
contexto inalterado e uma única implementação de lifecycle.

## Não-objetivos

Engine, runtime, scheduler, paralelismo, observabilidade automática, service
locator, DI container, adapters, transação distribuída, panic-safe cleanup e uma
nova álgebra funcional de valores continuam fora do módulo.

> Atualização: a exclusão absoluta de retry e timeout desta primeira álgebra foi
> superada por [ADR-011](adr-011-explicit-resilience-as-action-transformation.md).
> Permanecem válidas as exclusões de engine, runtime, background work,
> paralelismo implícito e execução durável.
