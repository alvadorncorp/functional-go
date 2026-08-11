# Resilience

`workflow/resilience` transforma uma `action.Action[T]` sem executá-la. A
política permanece visível no adapter que conhece o orçamento e a classe de
falha operacional.

## Retry seletivo

```go
policy := resilience.RetryPolicy{
    Attempts: 3,
    When: faults.IsUnavailable,
    Backoff: func(attempt int) time.Duration {
        return time.Duration(1<<(attempt-1)) * 100 * time.Millisecond
    },
}
program := resilience.Retry(policy, source)
value, err := action.Run(ctx, program)
```

`Attempts` inclui a chamada inicial. O argumento de Backoff é o número 1-based
da tentativa que falhou. `When` não roda depois da última falha. Espera zero é
imediata; espera negativa causa panic. Cancelamento antes da tentativa ou
durante a espera devolve o erro exato do contexto e o zero de T.

Retry não torna uma operação segura. Use-o somente quando o efeito é idempotente,
deduplicado, transacional no limite necessário, ou quando repetição parcial foi
modelada. O predicado de classificação pertence ao adapter, não ao pacote.

## Timeout cooperativo

```go
program := resilience.Timeout(3*time.Second, source)
value, err := action.Run(request.Context(), program)
```

Timeout deriva um contexto filho e chama a origem no mesmo goroutine. Ele é
cooperativo: se a origem ignora o contexto, não há preempção mágica.

## Ordem de composição

```go
totalBudget := resilience.Timeout(30*time.Second,
    resilience.Retry(policy, source),
)

perAttemptBudget := resilience.Retry(policy,
    resilience.Timeout(3*time.Second, source),
)
```

`Timeout(Retry(source))` limita o conjunto das tentativas e backoffs.
`Retry(Timeout(source))` renova o orçamento de cada tentativa. As duas formas
são deliberadamente diferentes.

## Transparência

A construção é lazy. O pacote não inicia goroutine, não recupera panic, não
embrulha erros, não registra métricas e não armazena estado global. Plan descreve
estrutura; Resilience transforma Action; somente Run executa.
