# Capacidades como structs de funções

Dependências são funções e structs pequenos montados na composition root
(`adr:workflow:003`). Exemplo checkout: `Orders{ Load func(...) action.Action[Order] }`.
Escritas só existem em `TxScope` dentro de `transaction.Within`.

Interfaces continuam corretas onde há polimorfismo genuíno; não são o
mecanismo mandatório de injeção. Nomes `Repository`, `Store`, `CommandBus`
são proibidos no núcleo de `workflow` e não devem ser reintroduzidos como
kit de aplicação.
