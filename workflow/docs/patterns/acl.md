# ACL entre bounded contexts

Uma função pura traduz os dados de um contexto para o vocabulário do outro.
Nenhum contexto precisa conhecer as regras de negócio do outro. Bounded
context não implica uma fronteira de transação distinta (`adr:workflow:006`).
Não extraia um tipo `ACL` para a biblioteca.
