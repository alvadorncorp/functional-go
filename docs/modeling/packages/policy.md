# Policy

**Import path:** `go.alvadorn.org/functional/modeling/policy`

`Policy[I,O]` é uma estratégia pura que deriva um resultado ou escolha de
política a partir de um contexto explícito. Ela responde **qual resultado ou
estratégia usar?**; não valida, executa efeito, busca dados nem conduz workflow.

## Escolha da operação

| Necessidade | Operação |
|---|---|
| resultado constante | `Constant` |
| transformar output | `Map` |
| adaptar input | `Contramap` |
| encadear `A -> B -> C` | `Compose` |
| calcular duas decisões independentes | `Zip` |
| escolher entre duas estratégias totais | `Choose` |
| tornar uma policy condicional | `When` |
| tentar alternativa binária | `OrElse` |
| primeira entre N alternativas | `FirstApplicable` |
| fornecer default | `Fallback` |

## Totalidade e inaplicabilidade

Use `Choose` quando exatamente um branch deve produzir `O`. O predicate é
avaliado uma vez e somente o branch escolhido roda. Use `When` quando a policy
pode não se aplicar; ela produz `Optional[O]`. `OrElse` combina duas formas
condicionais, `FirstApplicable` combina N e `Fallback` converte a seleção em uma
Policy total.

```go
vip := policy.When(isVIP, vipDiscount)
employee := policy.When(isEmployee, employeeDiscount)
volume := policy.When(isHighValue, volumeDiscount)

binary := policy.OrElse(vip, employee)
many := policy.FirstApplicable(vip, employee, volume)
discount := policy.Fallback(many, defaultDiscount)
_, _ = binary, discount
```

## Composição

`Compose(first, second)` executa first e passa seu output para second.

```go
creditPolicy := policy.Compose(customerTierPolicy, creditLimitPolicy)
```

`Zip` avalia left, depois right, depois combine sobre o mesmo input e produz o
tipo de domínio escolhido pela aplicação, sem tuple pública.

```go
pricing := policy.Zip(discountPolicy, shippingPolicy, func(discount Discount, shipping ShippingMethod) PricingDecision {
    return PricingDecision{Discount: discount, Shipping: shipping}
})
```

## Policy versus outros conceitos

| Conceito | Pergunta | Forma |
|---|---|---|
| Policy | qual estratégia/resultado usar? | `I -> O` |
| Specification | pertence a uma categoria de domínio? | `T -> bool` |
| Rule | satisfaz uma condição explicável? | `T -> error` |
| Requirement | a operação é permitida? | `S x I -> error` |
| Invariant | o estado resultante é válido? | `S -> error` |
| Change | como transformar o mesmo estado? | `S -> S` |

Uma Specification pode alimentar Choose, mas a aplicação faz a adaptação
explicitamente. Policy e Specification não importam uma à outra.

## Pureza, nil e zero values

Policies recebem todos os dados pelo input; banco, HTTP, broker, clock,
filesystem e context pertencem ao workflow/adapters. Zero output é um resultado
válido. A zero Policy é função nil quando invocada diretamente. Combinadores
rejeitam Policy, mapper, predicate, conditional ou combiner nil na construção
com os panics documentados no TDD.

Policy não possui FlatMap, engine, registry, dispatcher, métodos, marker types,
serviços, efeitos ou integração automática. Quando o vocabulário atual não
expressar um caso, a aplicação pode declarar uma Policy própria sobre um input
struct explícito.
