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
| `github.com/alvadorncorp/gopackages/workflow` | `Workflow[I,O]`, `From`, `Map`, `MapFallible`, `Then`, `Provide`, `Contramap` |
| `github.com/alvadorncorp/gopackages/workflow/action` | `Action[T]`, construção, transformação, combinação sequencial, observação, finalização, error reification e resource safety |
| `github.com/alvadorncorp/gopackages/workflow/plan` | descrição estrutural explícita, imutável e não executável |
| `github.com/alvadorncorp/gopackages/workflow/resilience` | `Retry`, `Timeout` e política de backoff como transformações explícitas de Action |
| `github.com/alvadorncorp/gopackages/workflow/transaction` | `Boundary[S]`, `Within` |

Plan describes; Workflow constructs; Action defers; Resilience transforms;
Boundary scopes; Run executes.

O núcleo depende apenas da biblioteca padrão, e um teste de arquitetura mantém
isso assim.

## Álgebra de composição

`Contramap` adapta input e `Provide` fixa o input sem executar efeitos.
`FlatMap` representa dependência; `Zip` preserva dois resultados independentes e
`Both` os combina, sempre em ordem sequencial. `Tap` observa sucesso, `Attempt`
transforma erro em dado, `Ensure` finaliza retornos normais e `Bracket` delimita
acquire/use/release sem possuir o recurso concreto.

## O que não é

Não é um motor de workflows, não é um framework DDD, não é um contêiner de
injeção de dependências e não é um sistema de efeitos. Não promete durabilidade,
paralelismo nem atomicidade distribuída. Retry e timeout existem apenas como
transformações explícitas e cooperativas em `workflow/resilience`.

## Documentação

Comece pela [visão geral](docs/overview.md), depois a
[arquitetura](docs/architecture.md). Índice completo em
[docs/README.md](docs/README.md), incluindo [Plan](docs/plan.md),
[Resilience](docs/resilience.md), ADRs e desenhos técnicos.

Os cenários executáveis de ponta a ponta vivem no módulo separado
`workflow/examples`: um fluxo de saque com outbox transacional e uma transferência
cross-context nas topologias de banco compartilhado e de bancos independentes.
