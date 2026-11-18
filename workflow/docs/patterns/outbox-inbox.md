# Outbox e inbox como composição da aplicação

Publicar num broker dentro da TX é o erro que o outbox existe para evitar
(`adr:workflow:transactional-outbox-as-application-composition`). `Outbox`, `OutboxRepository` e `Inbox` são nomes
proibidos no núcleo. A aplicação define `OutboxMessage` e `EventEncoder`.

Invariant: claim (se houver MessageID), writes de estado e appends de
outbox compartilham um `transaction.Boundary`. Eventos na ordem de
`Decision.Events()`; zero eventos ⇒ zero linhas.

Inbox (checkout `TxInbox.Claim`, crosscontext distributed): at-least-once
no consumidor; Duplicate curto-circuita sem reaplicar o domínio. Não
promova a primitivo até um terceiro domínio concordar no `MessageID`.

## Traverse para appends ordenados

O laço repetido de [load → decide → confirm](load-decide-confirm.md) —
codificar e fazer append de cada evento, parando no primeiro erro — é
`action.Traverse` sobre os eventos materializados:

```go
action.Traverse(slices.Collect(decision.Events()),
    func(e Event) action.Action[OutboxMessage] {
        return action.Tap(action.MapFallible(action.Pure(e), encode), outbox.Append)
    })
```

Encode e append ficam intercalados por evento, em ordem de índice; o erro do
encoder mantém identidade e os appends seguintes não rodam. Zero eventos ⇒
zero appends e slice vazio. A ordem vem da execução sequencial, não do
`iter.Seq`: `slices.Collect` só fixa o que cada `Run` percorre
(`adr:workflow:sequential-traverse-over-slices`).
