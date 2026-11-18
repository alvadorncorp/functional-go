# Operational events and integration messages

Operational events describe what happened during execution. Integration
messages define domain facts that other applications can consume.

Don't unify the contracts by default: the owners and the guarantees differ.
When the message must be atomic with the business change, write it to the
application outbox inside the same transaction
(`adr:workflow:transactional-outbox-as-application-composition`).
