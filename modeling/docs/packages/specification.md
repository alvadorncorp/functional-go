# Specification

**Import path:** `github.com/alvadorncorp/gopackages/modeling/specification`

`Specification[T]` é uma classificação booleana com significado explícito no
vocabulário do domínio. Ela responde **este valor pertence a esta categoria?**

```go
var IsActive specification.Specification[Customer] = func(customer Customer) bool {
    return customer.Active
}

if IsActive(customer) {
    // comportamento da aplicação
}
```

## Conceitos vizinhos

| Conceito | Pergunta | Forma |
|---|---|---|
| Predicate | condição booleana genérica | `T -> bool` |
| Specification | pertence à categoria de domínio? | `T -> bool` |
| Rule | satisfaz condição explicável? | `T -> error` |
| Requirement | operação é permitida? | `S x I -> error` |
| Invariant | estado resultante é válido? | `S -> error` |
| Policy | qual estratégia/resultado usar? | `I -> O` |

Predicate e Specification têm a mesma representação, mas não a mesma
responsabilidade. A identidade nominal de Specification comunica intenção de
domínio. Specification não retorna error nem mensagem; quando a falha precisa
ser explicada, use Rule/Requirement/Invariant conforme a pergunta.

## Composição

`All` faz AND left-to-right e para no primeiro false. `Any` faz OR left-to-right
e para no primeiro true. `All()` retorna true e `Any()` retorna false. `Not`
cria uma classificação derivada; não transforma classificação em validação.

```go
eligibleForCredit := specification.All(IsActive, HasVerifiedEmail, HasNoOutstandingDebt)
inactive := specification.Not(IsActive)
specialDiscount := specification.Any(IsVIP, IsEmployee, HasPromotionCode)
```

`Contramap` adapta a entrada mantendo a classificação original:

```go
isHighValueOrder := specification.Contramap(isHighValueMoney, func(order Order) Money {
    return order.Total
})
```

Para mais de um valor, declare um contexto do domínio e classifique esse valor:

```go
type TransferContext struct {
    Account Account
    Amount  Money
}

var RequiresManualReview specification.Specification[TransferContext]
```

## Specification e Policy

A aplicação pode usar uma Specification como condição de uma Policy, mantendo a
adaptação explícita:

```go
discount := policy.Choose(
    func(context PricingContext) bool { return IsVIP(context) },
    vipDiscount,
    regularDiscount,
)
```

Não existe dependência Policy -> Specification nem Specification -> Policy.

## Superfície e nil

A superfície é exatamente `Specification`, `All`, `Any`, `Not` e `Contramap`.
`Of` foi rejeitado porque atribuição/conversão direta é suficiente; `Xor` foi
rejeitado por falta de caso concreto. Filter, Validate, Explain, Reason, Error e
Message pertencem a outras responsabilidades.

A zero Specification é função nil ao ser chamada. Combinadores rejeitam
Specification nil eager; Contramap também rejeita mapper nil. Nil nunca é true,
false ou skip. O package não tem imports de produção, efeitos, runtime,
reflection, registry, persistence ou infraestrutura.
