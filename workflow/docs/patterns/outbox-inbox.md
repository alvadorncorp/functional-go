# Outbox e inbox como composição da aplicação

Publicar num broker dentro da TX é o erro que o outbox existe para evitar
(`adr:workflow:008`). `Outbox`, `OutboxRepository` e `Inbox` são nomes
proibidos no núcleo. A aplicação define `OutboxMessage` e `EventEncoder`.

Invariant: claim (se houver MessageID), writes de estado e appends de
outbox compartilham um `transaction.Boundary`. Eventos na ordem de
`Decision.Events()`; zero eventos ⇒ zero linhas.

Inbox (checkout `TxInbox.Claim`, crosscontext distributed): at-least-once
no consumidor; Duplicate curto-circuita sem reaplicar o domínio. Não
promova a primitivo até um terceiro domínio concordar no `MessageID`.
