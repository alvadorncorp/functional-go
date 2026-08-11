# workflow

`workflow` oferece primitivos pequenos e type-safe para construir, compor e
executar workflows reutilizáveis com efeitos em Go.

Workflows descrevem a orquestração da aplicação; regras de domínio permanecem no
domínio, e a infraestrutura permanece atrás de capacidades definidas pela
aplicação.

```go
import (
    "github.com/alvadorncorp/gopackages/workflow"
    "github.com/alvadorncorp/gopackages/workflow/action"
)

type Users struct {
    Find func(UserID) action.Action[User]
}

func LoadUser(users Users) workflow.Workflow[UserID, User] {
    return workflow.From(users.Find)
}

loadUser := LoadUser(users)

program := loadUser(id)
// No effect has executed yet.

user, err := action.Run(ctx, program)
```

## Os três momentos

Dependências entram na construção, a entrada entra na aplicação e o contexto
entra na execução. Construir e compor workflows não realiza I/O; apenas
`action.Run` inicia qualquer coisa.

## Pacotes

| Caminho de importação | Conteúdo |
|---|---|
| `github.com/alvadorncorp/gopackages/workflow` | `Workflow[I,O]`, `From`, `Map`, `MapFallible`, `Then` |
| `github.com/alvadorncorp/gopackages/workflow/action` | `Action[T]`, `Run`, `Do`, `Pure`, `Fail`, `Map`, `MapFallible`, `FlatMap`, `Recover` |
| `github.com/alvadorncorp/gopackages/workflow/transaction` | `Boundary[S]`, `Within` |

O núcleo depende apenas da biblioteca padrão, e um teste de arquitetura mantém
isso assim.

## O que não é

Não é um motor de workflows, não é um framework DDD, não é um contêiner de
injeção de dependências e não é um sistema de efeitos. Não promete durabilidade,
retry, timeout, paralelismo nem atomicidade distribuída.

## Documentação

Comece pela [visão geral](docs/overview.md), depois a
[arquitetura](docs/architecture.md). Índice completo em
[docs/README.md](docs/README.md), incluindo oito ADRs e cinco desenhos técnicos.

Os cenários executáveis de ponta a ponta vivem no módulo separado
`workflow/examples`: um fluxo de saque com outbox transacional e uma transferência
cross-context nas topologias de banco compartilhado e de bancos independentes.
