# Eventos operacionais e mensagens de integração

Eventos operacionais descrevem o que aconteceu durante a execução. Mensagens
de integração definem fatos do domínio que outras aplicações podem consumir.

Não unifique os contratos por padrão: os donos e as garantias são diferentes.
Quando a mensagem precisa ser atômica com a alteração de negócio, grave-a no
outbox da aplicação dentro da mesma transação (`adr:workflow:008`).
