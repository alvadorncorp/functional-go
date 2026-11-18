# Load → decide → confirm

Forma canônica (`tdd:workflow:modeling-integration`): carga com efeito,
função de domínio pura, confirmação com efeito. `workflow ↛ modeling`.

Dois tempos de `Decide`, ambos legítimos:

| | Decide | TX se o domínio recusar |
|---|---|---|
| withdraw, createorder | fora da TX | não abre |
| checkout | dentro da TX (após claim) | TX já aberta |

Crosscontext: duas decisões puras, depois uma TX compartilhada **ou** TX
local + relay. Não extraia `CommitDecision` até três domínios concordarem
na forma (`adr:workflow:modeling-integration-boundary` — hoje não concordam).

O fragmento que se repete é iterar `decision.Events()` e fazer append
**sequencial** no outbox da aplicação (ver outbox-inbox.md).
