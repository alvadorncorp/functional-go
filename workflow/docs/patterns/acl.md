# ACL entre bounded contexts

`examples/workflow/crosscontext/acl` importa os dois contextos e expõe uma
função pura de tradução. Banking não aprende livro; Ledger não aprende
saldo. Bounded context ≠ fronteira de transação (`adr:workflow:006`).
Não extraia um tipo `ACL` para a biblioteca.
