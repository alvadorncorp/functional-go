# Composição

Workflows se constroem em um vocabulário:

```text
LoadUser  +  RequireActive        =  LoadActiveUser
LoadActiveUser  +  LoadPermissions =  AuthorizedUser
AuthorizedUser  +  ChangeEmail     =  ChangeEmailWorkflow
```

`LoadActiveUser` é um `Workflow` ordinário. Nada o marca como composto, e
workflows maiores não conseguem distinguir — que é exatamente por que o
vocabulário cresce sem o consumidor aprender uma árvore interna.

```go
loadActiveUser := workflow.Then(loadUser, requireActive)
```

## A composição é binária

Os genéricos de Go não preservam um tipo intermediário que muda ao longo de uma
lista variádica, então `Then` recebe dois workflows. Encadeie-os:

```go
ab  := workflow.Then(a, b)
abc := workflow.Then(ab, c)
```

## Nem tudo deve ser um workflow

Uma função pura permanece função. `func NormalizeEmail(string) string` não ganha
nada virando um `Workflow`. Use um quando houver efeito, orquestração
reutilizável, captura de dependência, composição entre capacidades ou ciclo de
vida de execução. Uniformidade não vale perder significado.

Da mesma forma, uma função de domínio pura que retorna `modeling.Decision` deve
permanecer função de domínio pura; `MapFallible` existe para que ela entre em um
workflow sem se disfarçar de efeito.

## Adaptar e fornecer input

`Contramap(wf, mapper)` transforma `Workflow[I,O]` em `Workflow[A,O]` por uma
projeção `A → I`. O mapper roda quando o workflow é aplicado. `Provide(wf, input)`
elimina a entrada e retorna `Action[O]`; não chama `Run`.

## Dependência, independência e paralelismo

| Operação | Relação | Resultado |
|---|---|---|
| `FlatMap` | a segunda Action depende do primeiro valor | `Action[B]` |
| `Zip` | resultados independentes, execução sequencial | `Action[Pair[A,B]]` |
| `Both` | resultados independentes, execução sequencial | `Action[C]` |

Independência descreve data flow, não concorrência. A ordem continua esquerda →
direita, sem goroutines.
