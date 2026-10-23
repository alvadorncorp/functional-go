# Resilience

`workflow/resilience` transforma uma `action.Action[T]` sem executá-la. A
política permanece visível no adapter que conhece o orçamento e a classe de
falha operacional.

## Retry seletivo

```go
policy := resilience.RetryPolicy{
    Attempts: 3,
    When: faults.IsUnavailable,
    Backoff: exponentialBackoff,
}
program := resilience.Retry(policy, source)
value, err := action.Run(ctx, program)
```

O adapter pode limitar o backoff antes do shift, usando `math/bits` da stdlib:

```go
func exponentialBackoff(attempt int) time.Duration {
    const baseDelay = 100 * time.Millisecond
    const maxDelay = 2 * time.Second
    if attempt <= 1 {
        return baseDelay
    }
    if attempt-1 >= bits.Len64(uint64(maxDelay/baseDelay)) {
        return maxDelay
    }
    return baseDelay << (attempt - 1)
}
```

O limite antes do shift impede overflow mesmo com um número alto de tentativas.

`Attempts` inclui a chamada inicial. O argumento de Backoff é o número 1-based
da tentativa que falhou. `When` não roda depois da última falha. Backoff nil ou
espera zero são imediatos; espera negativa causa panic. Cancelamento antes da
tentativa ou durante a espera devolve o erro exato do contexto e o zero de T.

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

perAttemptPolicy := policy
perAttemptPolicy.When = func(err error) bool {
    return faults.IsUnavailable(err) || errors.Is(err, context.DeadlineExceeded)
}
perAttemptBudget := resilience.Retry(perAttemptPolicy,
    resilience.Timeout(3*time.Second, source),
)
```

`Timeout(Retry(source))` limita o conjunto das tentativas e backoffs.
`Retry(Timeout(source))` renova o orçamento de cada tentativa; por isso o segundo
predicado também aceita `context.DeadlineExceeded` do contexto filho. O
`ctx.Err()` antes de cada tentativa e a espera cancelável de Retry continuam
respeitando o contexto pai: se o orçamento total dele acabar, não há nova
tentativa. O timeout permanece cooperativo se a origem ignorar o contexto.

## Transparência

A construção é lazy. O pacote não inicia goroutine, não recupera panic, não
embrulha erros, não registra métricas e não armazena estado global. Plan descreve
estrutura; Resilience transforma Action; somente Run executa.
