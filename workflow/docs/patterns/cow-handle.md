# Handle COW para aggregates grandes

`hugeParam` (gocritic, 80 bytes) é um guarda-corpo sobre valores que a
**aplicação** copia por assinatura — não um motivo para ponteirizar
`Decision`, `Reducer` ou `Workflow` (`adr:modeling:streaming-replay-project-seq`).

Padrão (checkout `Order`, withdraw `Account`): um struct `Order { p *orderPayload }`
com API por valor; payload imutável após publicar.

Getters no zero (`p == nil`) devolvem zeros; edições rejeitam. Coleções
variáveis vivem em `functional/immutable` **dentro** do payload.

Não extraia `Handle[T]`. Ver `functional/docs/patterns/cow-handles.md`.
