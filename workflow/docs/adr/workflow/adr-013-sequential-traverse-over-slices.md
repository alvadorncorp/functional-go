---
type: adr
id: adr:workflow:013
context: workflow
title: Traverse sequencial sobre slices
status: Aceita
tags: [workflow, action, composition, traverse, order, outbox]
deciders: [alvadorn]
related: [adr:workflow:002, adr:workflow:007, adr:workflow:009, adr:modeling:010, adr:functional:008, tdd:workflow:action]
date: 2026-10-20
---
# ADR-013: Traverse sequencial sobre slices

## Contexto e problema

A confirmação de uma decisão percorre os eventos emitidos e, para cada um,
executa uma Action (codificar e fazer append no outbox), parando no primeiro
erro. A álgebra de `action` não tem combinador para "uma Action por item, em
ordem", então cada aplicação escreve o mesmo laço manual com `action.Run`
dentro de uma Action externa.

O guia de padrões já nomeia a repetição
(`docs/patterns/load-decide-confirm.md:17`): iterar `decision.Events()` e fazer
append **sequencial** no outbox da aplicação. Os laços são idênticos em ordem,
curto-circuito e contexto, mas cada cópia reimplementa essas regras sem teste
próprio.

ADR-007 recusa famílias de monads e emulação higher-kinded; ADR-009 recusa
paralelismo e uma nova álgebra funcional de valores. Um `Traverse` precisa caber
dentro desses limites.

## Drivers da decisão

- ordem de efeitos observável e igual à ordem dos itens;
- `Action` continua adiada e reexecutável (ADR-002): cada `Run` refaz o mesmo
  trabalho a partir da mesma entrada;
- mesmas convenções de `FlatMap`, `Zip` e `Both`: curto-circuito, erro com
  identidade preservada, panic na construção para argumento nil e panic na
  execução para Action nil;
- `action` continua importando exatamente `context` e `errors`;
- nenhuma goroutine, nenhum `recover` e nenhuma política de contexto escondida.

## Decisão

`action` recebe:

```go
func Traverse[A, B any](items []A, binder func(A) Action[B]) Action[[]B]
```

`Traverse` é monomórfico: opera sobre `[]A` e `Action`, sem interface
`Traversable`, sem parâmetro de tipo de container e sem emulação higher-kinded.
Isso reconcilia o não-objetivo de ADR-007. Ele é estritamente sequencial e não
cria álgebra de valores, o que reconcilia o não-objetivo de ADR-009: é
composição de efeitos sobre a mesma `Action` que `Zip` e `Both` já compõem.

Semântica:

- os itens são vinculados e executados em ordem de índice, intercalados: o
  binder do item `i+1` só é chamado depois que a Action do item `i` teve
  sucesso, e toda Action recebe o contexto dado a `Run`;
- nada executa na construção; cada `Run` percorre os itens de novo;
- a construção copia `items` com `append([]A(nil), items...)`; a cópia é rasa,
  então elementos ponteiro continuam compartilhando o alvo;
- o sucesso devolve um slice novo, não nil, com `len(items)` valores; entrada nil
  ou vazia devolve slice vazio não nil sem chamar o binder;
- a primeira falha devolve `(nil, err)` com o erro inalterado, e o binder não é
  chamado para os itens restantes;
- não há checagem de `ctx` entre itens: o contexto, mesmo cancelado, chega
  inalterado às Actions, que observam o cancelamento como qualquer outra;
- binder nil gera panic `workflow/action: nil binder` na construção, inclusive
  com `items` nil; binder que devolve Action nil gera panic
  `workflow/action: nil action` na execução. Nenhuma constante de panic nova.

## Alternativas consideradas

- **Entrada `iter.Seq[A]`.** `adr:modeling:010` adotou `iter.Seq` em `Replay` e
  `Project` para evitar materialização num fold de passagem única. Rejeitado
  aqui: uma `Action` é reexecutável e um `iter.Seq` não promete produzir os
  mesmos itens em uma segunda passagem. A cópia em slice na construção fixa o
  que cada `Run` executa; a reexecutabilidade vence a economia de uma alocação.
  O caller materializa explicitamente os eventos de `decision.Events()` antes de
  chamar `Traverse`, e a ordem continua garantida porque os appends são
  sequenciais em ordem de índice.
- **Nomes `ForEach` ou `Each`.** Rejeitados: sugerem iteração só por efeito,
  descartando valores, enquanto o combinador coleta um resultado por item.
  `Traverse` nomeia exatamente "mapear com efeito e coletar".
- **Acumular falhas com `errors.Join`.** Rejeitado: exigiria executar os itens
  seguintes depois de uma falha, o que num outbox significa gravar mensagens
  depois de um erro dentro da transação. Acumulação de erros pertence a
  `functional/values/validated`; aqui o primeiro erro mantém identidade.
- **Execução paralela.** Rejeitado pelo mesmo motivo de ADR-009 para `Zip` e
  `Both`: exigiria goroutines, sincronização e política de cancelamento, e a
  ordem dos appends faz parte do que o consumidor observa.
- **Checar `ctx.Err()` entre itens.** Rejeitado: seria política de contexto
  escondida no combinador. ADR-009 fixa que o mesmo contexto, mesmo cancelado,
  chega a toda etapa; quem decide parar é a Action ou o adapter.
- **Copiar com `slices.Clone`.** Rejeitado: adicionaria `slices` aos imports de
  `action`, que o gate de arquitetura mantém exatamente em `context` e `errors`.
  `append([]A(nil), items...)` é equivalente.
- **Colocar `Traverse` em `functional`.** Rejeitado: `functional` não tem sistema
  de efeitos (ADR-007), e `adr:functional:008` mantém suas primitivas como folhas
  puras sobre valores. `Action` carrega `context.Context` e semântica de
  execução, então o combinador vive ao lado dela.

## Consequências

A superfície de `action` cresce de forma aditiva em uma função, fixada pelos
gates de superfície e de contrato e exercitada por exemplo compilado. O perfil de
alocação é uma cópia de `items` na construção, um slice de resultado por `Run` e
as alocações do próprio binder. Callers com `iter.Seq` pagam uma materialização
explícita; a invariante de ordem continua valendo porque a execução é
sequencial. A cópia rasa e a ausência de checagem de contexto ficam
documentadas no contrato. Emenda (2026-10-23): o `appendOutbox` do withdraw
passou a usar `Traverse` sobre os eventos materializados, como o checkout. Os
laços `for` com `action.Run` de crosscontext permanecem de propósito, por
motivos distintos. Em `distributed`, `appendIntegrationMessages` filtra por
evento (só `TransferOccurred` vai ao outbox), o que torna `Traverse` desajeitado
ali. Em `shared`, `appendAll` não filtra: grava uma mensagem por evento de duas
decisões distintas, uma depois da outra, e o laço imperativo é Go idiomático.
Não há adoção pendente.

## Não-objetivos

`Traverse` paralelo, variantes sobre `iter.Seq` ou mapas, acumulação de erros,
`Sequence` sobre `[]Action[T]`, typeclass `Traversable` e retry ou timeout
embutidos continuam fora do módulo.
