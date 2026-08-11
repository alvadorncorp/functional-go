---
type: adr
id: adr:modeling:009
context: modeling
title: Álgebra fechada de composição de Policy
status: Aceita
tags: [modeling, policy, composition, optional, compatibility]
deciders: [alvadorn]
related: [adr:modeling:005, adr:modeling:008, tdd:modeling:policy]
date: 2026-08-11
---
# ADR-009: Álgebra fechada de composição de Policy

## Contexto e problema

ADR-005 estabeleceu Policy como estratégia pura `I -> O`, sete exports,
inaplicabilidade por Optional e panics eager para operandos nil. Faltavam
composição sequencial, combinação de decisões independentes, escolha total e
fallback condicional binário. A linha estável precisa completar esse vocabulário
sem transformar Policy em rules engine ou mônada.

## Direcionadores

- preservar os sete contratos e duas dependências existentes;
- tornar ordem e short-circuit parte observável do contrato;
- manter totalidade separada de inaplicabilidade;
- evitar tuple pública, FlatMap, runtime, bridge ou nova abstração nominal;
- congelar a menor superfície semanticamente completa.

## Alternativas consideradas

### FlatMap e álgebra monádica

Rejeitada: mistura o input original com decisões subsequentes e cresce a
abstração sem um significado de domínio único.

### Tuple pública para Zip

Rejeitada: vazaria estrutura de `functional`; um combiner explícito produz o
tipo de domínio que a aplicação realmente quer.

### Conditional nominal

Rejeitada novamente: `func(I) optional.Optional[O]` continua suficiente e
compatível com When, FirstApplicable e Fallback.

### Quatro combinadores fechados (escolhida)

Escolhida: adicionar `Compose`, `Zip`, `Choose` e `OrElse`.

## Decisão

`Compose(first, second)` avalia `first` e alimenta seu output em `second`.
`Zip(left, right, combine)` avalia estritamente left, depois right, depois
combine. `Choose(predicate, whenTrue, whenFalse)` avalia o predicate uma vez e
somente o branch selecionado. `OrElse(left, right)` é a composição binária de
policies condicionais; `FirstApplicable` permanece a forma N-ária.

Choose usa `predicate.Predicate[I]`, igual a When. A aplicação adapta uma
Specification por closure ou conversão explícita, tornando a fronteira
semântica visível sem dependência entre siblings. OrElse preserva Optional e
short-circuit left-to-right.

Compose/Zip/Choose reutilizam o panic existente para Policy nil; Choose
reutiliza o panic de Predicate nil; OrElse reutiliza o panic de Conditional
nil. Zip adiciona `modeling/policy: nil combiner`. Nenhuma dependência nova é
necessária e os sete símbolos anteriores ficam inalterados.

## Compatibilidade e rollback

A mudança é aditiva e amplia a superfície de sete para onze símbolos. Antes de
publicação, cada slice pode ser revertido. Depois de publicação, remoção,
assinatura, ordem, short-circuit, zero-output ou nil semantics dos onze símbolos
é breaking e requer nova decisão e SemVer. Não há publicação neste trabalho.

## Consequências

Policy fica feature-complete para a estabilização: Map transforma output,
Contramap adapta input, Compose encadeia, Zip combina, Choose seleciona uma
estratégia total, When introduz inaplicabilidade, OrElse tenta alternativa,
FirstApplicable busca entre N e Fallback totaliza. Novas abstrações exigem um
problema de domínio não expressável por esse vocabulário.
