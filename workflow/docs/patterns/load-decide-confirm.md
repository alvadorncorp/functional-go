# Load → decide → confirm

Canonical form (`tdd:workflow:modeling-integration`): load with effect, pure
domain function, confirm with effect. `workflow ↛ modeling`.

Two `Decide` timings, both legitimate:

| | Decide | TX if the domain refuses |
|---|---|---|
| withdraw, createorder | outside the TX | no TX opens |
| checkout | inside the TX (after claim) | TX already open |

Crosscontext: two pure decisions, then one shared TX **or** a local TX +
relay. Don't extract `CommitDecision` until three domains agree on the
shape (`adr:workflow:modeling-integration-boundary`; today they don't).

The repeated fragment is iterating `decision.Events()` and appending
**sequentially** to the application outbox (see outbox-inbox.md).
