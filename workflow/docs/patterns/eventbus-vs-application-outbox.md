# eventbus/outbox vs outbox da application

`eventbus/outbox` (`adr:eventbus:001`) é publicação durável de
`eventbus.Event` com TX ambiente — observabilidade/integração do **bus**.
O outbox do use case (`adr:workflow:008`) é mensagem de integração do
domínio da aplicação (`OutboxMessage` + encoder).

Não unifique por padrão: contratos e donos diferem. Um adapter pode
persistir ambos na mesma TX física sem fundir as APIs.

Evidência: `examples/workflow/checkout` e `examples/workflow/withdraw` usam o outbox da aplicação; não unifique com `eventbus/outbox`.
