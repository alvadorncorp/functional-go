---
type: adr
id: adr:functional:012
context: functional
title: Interop de Result com o par (T, error) e regra de ordem do Match
status: Aceita
tags: [functional, values, result, error-handling, public-api]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:007, adr:workflow:007]
date: 2026-10-20
---
# ADR-012: Interop de Result com o par (T, error) e regra de ordem do Match

## Contexto e Problema

`values/result` modela sucesso e falha, mas toda API Go devolve `(T, error)`.
Sem ponte, cada consumidor reescreve `if err != nil { return result.Err... }`
na entrada e pares `Get`/`ErrValue` na saída; `examples/workflow/checkout`
mantém um helper `unwrap` local só para isso. Além disso, só `either` tem um
eliminador (`Match`); `result`, `optional` e `validated` obrigam o caller a
ramificar à mão sobre `Get`.

A documentação da família (`values.md`, `values-result.md`) excluía
"conversões" sem distinguir conversão entre tipos da família de interop com o
protocolo de erro do Go. O ADR-007 de `workflow` já rejeitou pontes
automáticas `FromResult`/`ToResult` dentro de `workflow`.

## Fatores de Decisão

- Manter `values/result` só com stdlib e sem novos imports (ADR-003, ADR-007).
- Nunca transformar uma falha em sucesso silenciosamente.
- Uma única regra de ordem para todos os `Match` da família.
- Custo zero de alocação no wrapper, como no resto da família.

## Opções Consideradas

### Opção 1: Métodos `r.ToPair()` e `FromPair` como construtor

Rejeitada: um método de `Result[T, E]` não consegue fixar `E = error`.
Funções de pacote fixam o tipo no parâmetro.

### Opção 2: `ToPair` devolve `(zero, nil)` para `Err(nil)`

Rejeitada: um `Result[T, error]` zero (por exemplo, um campo não
inicializado) viraria sucesso. É o comportamento do `unwrap` do checkout e
esconde falhas.

### Opção 3: `ToPair` devolve um erro sentinela para `Err(nil)`

Rejeitada: inventa uma falha que o caller não produziu e cria um novo símbolo
exportado. O estado é um erro de programação, como `MustGet` em `Err`.

### Opção 4: Conversões entre tipos da família (Result ↔ Either ↔ Validated)

Rejeitada e mantida fora de escopo: os nomes são compromissos semânticos
(ADR-007). `FromPair`/`ToPair` não convertem entre tipos da família; ligam
`Result` ao protocolo `(T, error)` da linguagem.

### Opção 5: `FromPair`, `ToPair` e `Match` como funções de pacote (escolhida)

## Decisão

Adicionar a `values/result`:

```go
func FromPair[T any](value T, err error) Result[T, error]
func ToPair[T any](r Result[T, error]) (T, error)
func Match[T, E, R any](r Result[T, E], errFn func(E) R, okFn func(T) R) R
```

**Semântica de `FromPair`.** Só verifica `err != nil`: `nil` gera `Ok(value)`
(inclusive valor zero); não-nil gera `Err(err)` e descarta `value`. Uma
interface que guarda ponteiro nil tipado (`error((*E)(nil))`) é não-nil e
gera `Err`. A documentação avisa dois casos inadequados: funções cujo último
resultado é um tipo de erro concreto ponteiro (`FromPair(f())` converte o nil
concreto em `error` não-nil e transforma sucesso em falha, sem erro de
compilação) e APIs `(n, err)` com valor parcial, que é descartado.

**Semântica de `ToPair`.** `Ok(v)` vira `(v, nil)`; `Err(e)` com `e` não-nil
vira `(zero, e)`. `Err(nil)`, incluindo o `Result[T, error]` zero, entra em
pânico com exatamente:

```
functional/values/result: ToPair on Err with nil error
```

O prefixo segue o pânico existente de `MustGet`.

**Regra de ordem do `Match`.** O handler do ramo de valor zero vem primeiro,
como em `either.Match(e, leftFn, rightFn)` e na tabela de valores zero da
família: `result.Match(r, errFn, okFn)`, `optional.Match(o, noneFn, someFn)` e
`validated.Match(v, invalidFn, validFn)`. A ordem sucesso-primeiro (como
`policy.Choose` de `modeling`) foi descartada para manter uma regra única.
O `Match` de `optional` e `validated` segue esta regra quando for adicionado.

**Handlers nil.** Como no resto da família, `Match` chama os handlers
diretamente, sem checagem de nil; exatamente um handler roda por chamada.

**Fronteira com `workflow`.** O ADR-007 de `workflow` continua válido: a
ponte `FromResult`/`ToResult` dentro de `workflow` segue rejeitada.
`FromPair`/`ToPair` vivem em `functional`, ligam `Result` ao `(T, error)` do
Go e não criam aresta `workflow` → `functional`; um consumidor escreve
`result.FromPair(act(ctx))` sem que `workflow` importe nada.

## Consequências

**Positivas:** consumidores apagam helpers locais como `unwrap`; a família
ganha um eliminador com ordem previsível; `FromPair`, `ToPair` e `Match` não
alocam no wrapper.

**Negativas:** a ordem do `Match` e a mensagem de pânico de `ToPair` viram
contrato permanente após o release. Um `Match` com handlers trocados só compila
para `Result[X, X]`. A armadilha do tipo de erro concreto não é detectável em
compilação; fica documentada e coberta por teste.

**Verificação:** testes de tabela e leis de ida e volta em
`values/result`, testes de alocação zero, `TestValuesPublicAPI` no archtest,
`go test ./...` e lint.
