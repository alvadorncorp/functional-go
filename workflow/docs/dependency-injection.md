# Injeção de dependências

Uma capacidade é um pequeno struct de funções:

```go
type Users struct {
    Load func(UserID) action.Action[User]
    Save func(User) action.Action[struct{}]
}
```

Um construtor de workflow recebe o que precisa e devolve um valor:

```go
func ChangeEmail(users Users, clock Clock) workflow.Workflow[ChangeEmailCommand, User]
```

A composition root é o único lugar que conhece várias implementações concretas de
uma vez:

```text
PostgreSQL   SMTP   Clock
        │
        ▼
   capabilities
        │
        ▼
    workflows
        │
        ▼
   HTTP adapters
```

## Por que isto não é um service locator

Nada é registrado e nada é resolvido. Uma dependência ausente é erro de
compilação, não pânico em runtime. O grafo de dependências de uma operação é
legível a partir da assinatura do construtor, e um workflow construído com um
conjunto de capacidades continua usá-las — não há resolução tardia que as
substitua.

## Nem toda dependência é um efeito

Um relógio e um gerador de identificadores não precisam de contexto, então são
funções simples:

```go
type Clock func() time.Time
type NewMessageID func() MessageID
```

Use a forma mais simples que preserve testabilidade. Embrulhar `Clock` em uma
`Action` adicionaria cerimônia e nenhuma informação.

## Testar sem framework

Um literal de struct com closures é o dublê:

```go
users := Users{
    Load: func(id UserID) action.Action[User] { return action.Pure(existing) },
    Save: func(u User) action.Action[struct{}] { saved = u; return action.Pure(struct{}{}) },
}
```

## `Provide` fornece input, não dependência

`workflow.Provide(loadUser, id)` aplica um input já conhecido a um workflow e
retorna uma Action. Não registra, busca nem substitui capacidades. Dependências
continuam entrando explicitamente no construtor; `Inject`, `Resolve`, `Container`
e `Registry` permanecem proibidos.
