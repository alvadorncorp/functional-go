# Architecture overview

`modeling` fornece um metamodelo, não um domínio de aplicação:

```text
application domain
      ├── specializes → modeling root facade
      │                     ↓ delegates privately
      │               modeling/internal/decision
      │                     ↓ uses
      │               functional/immutable
      ├── specializes → modeling/change
      │                     ↓ privately adapts
      │               functional/primitives/function
      ├── specializes → modeling/policy
      │                     ↓ uses
      │               functional/primitives/predicate
      │               functional/values/optional
      ├── specializes → modeling/reducer
      │                     ↓ uses
      │               stdlib iter
      ├── specializes → modeling/projection
      │                     ↓ uses
      │               stdlib iter
      └── specializes → modeling/specification
                            (zero production imports)
```

Aplicações importam o pacote raiz para Decision, Transition, Invariant, Rule e
Requirement. Importam `modeling/change` para evolução nomeada de estado,
`modeling/policy` para estratégias derivadas de contexto,
`modeling/reducer` para replay fato→estado, `modeling/projection` para
projeção fato→representação derivada e `modeling/specification` para
classificação booleana nominal. A raiz não reexporta Change, Policy, Reducer,
Replay, Projection, Project nem Specification. O pacote interno de Decision
permanece fronteira de implementação, não outra API pública.

## Semântica estrutural

Value, Identity, Entity, Aggregate e Model permanecem tipos Go da aplicação.
O módulo não exporta marker interfaces, embeddings obrigatórios, tipos base
nem igualdade por reflection. Um aggregate é fronteira de consistência, não
linha de database nem registro de repositório.

## Primitivas comportamentais

O contrato público da raiz contém:

- `Decision[S,E]` com zero value válido, construtor, observações e funções de
  composição tipadas;
- `Rule[T]` para condição explicativa sobre um valor, composta por `AllRules`
  ordenado ou `AnyRule` não vazio;
- `Requirement[S,I]` para pré-condição operacional sobre estado + input,
  composta por `AllRequirements` ou `AnyRequirement` não vazio;
- `Transition[S,I,O]` e `FallibleTransition[S,I,O]` para computação de
  próximo estado;
- `Invariant[S]` para validade do estado resultante, com composição fail-fast
  ou acumuladora de erros.

Capabilities em subpacotes:

- `reducer.Reducer[S,E]` e `reducer.Replay` — fold esquerdo síncrono sobre
  `iter.Seq[E]` (ver seção Reducer);
- `projection.Projection[P,E]` e `projection.Project` — fold correspondente
  para representações derivadas sobre `iter.Seq[E]`;
- `change.Change[S]` com Identity, Compose, Pipe, Apply e ApplyAll;
- `policy.Policy[I,O]` e combinadores de composição;
- `specification.Specification[T]` com All/Any/Not/Contramap.

Rule e Invariant permanecem tipos nominais distintos mesmo ambos devolvendo
error. Requirement responde se um input pode ser aplicado ao estado corrente;
não é hook de Transition. Change nem decide aplicabilidade nem valida ou emite
eventos. Funções ordinárias permanecem atribuíveis e erros de domínio são da
aplicação.

## Capability Policy

`modeling/policy` representa `I -> O`: estratégia pura que deriva um resultado
de domínio a partir de input explícito. Map transforma output, Contramap
adapta input, Compose encadeia decisões, Zip combina duas decisões no mesmo
input, Choose seleciona uma estratégia total, e When/OrElse/FirstApplicable/
Fallback modelam aplicabilidade ordenada. Policy não valida e não executa
efeito.

`modeling/specification` representa `T -> bool` nominal: classificação
booleana de domínio. All/Any avaliam left-to-right com short-circuit, Not
deriva a categoria complementar e Contramap adapta input. O pacote tem zero
imports de produção e é independente de Policy e de todo sibling.

Policy pergunta qual resultado ou estratégia usar. Specification pergunta se
um valor pertence a uma categoria de domínio. Rule explica condição falha,
Requirement autoriza operação, Invariant valida estado resultante e Change
realiza `S -> S`. Assinaturas subjacentes iguais nunca apagam essas
responsabilidades semânticas.

## Reducer e Replay

`Reducer[S,E]` é um `func(S,E) S` ordinário: um fato já ocorrido materializa-se
no próximo estado. Distingue-se deliberadamente de Transition (intenção/input),
de Decision (estado + eventos emitidos) e de Policy (estratégia a partir de
contexto).

`reducer.Replay(initial, events, reduce)` aplica cada evento de um
`iter.Seq[E]` exatamente uma vez, na ordem de encontro, em passagem única.
Replay vazio devolve o estado inicial; concatenar sequências (quando a
aplicação define `A ++ B`) é equivalente a reaplicar a esquerda e depois a
direita. A ponte com Decision é direta:

```go
state := reducer.Replay(previous, decision.Events(), reduce)
```

Não é necessário `immutable.CollectList`. Quem precisar de multi-pass
materializa explicitamente antes de chamar Replay.

Se esse Replay deve reproduzir `decision.State()` é lei da aplicação,
verificada nos testes dela: ver [Lei decide/evolve](patterns/decide-evolve-law.md).
A biblioteca não faz essa comparação. O padrão
[Decider](patterns/decider.md) mostra como a aplicação compõe Transition,
Decision e Reducer em `decide`/`evolve` sem tipo de biblioteca.

Produção de `modeling/reducer` depende somente do stdlib (`iter`). Replay é um
fold puro, não um runtime de Event Sourcing: não carrega streams, não
persiste, não serializa versões, não cria snapshots, não registra handlers e
não compara estado de Decision com estado reduzido.

## Projection e Project

`Projection[P,E]` é um `func(P,E) P` ordinário que deriva o próximo read
model, sumário, índice, estatística ou outra representação application-owned
a partir de um fato. `projection.Project(initial, events, project)` aplica a
mesma operação a um `iter.Seq[E]` na ordem de encontro, de passagem única, de
modo que rebuild em lote e atualização incremental compartilham uma função
semântica.

Projection distingue-se de Reducer. Reducer explica o que os fatos significam
para o estado de domínio; Projection explica o que significam para uma
representação derivada. Os dois pacotes são siblings, não se importam nem
importam a raiz de modeling, e consomem independentemente eventos da
aplicação.

Ponte com Decision:

```go
view := projection.Project(empty, decision.Events(), project)
```

Entrada vazia devolve `initial`; a lei de concatenação é
`Project(p,A++B,f) = Project(Project(p,A,f),B,f)` quando a aplicação define a
concatenação. Determinismo e imutabilidade superficial permanecem
responsabilidade da função fornecida.

Produção de `modeling/projection` depende somente do stdlib (`iter`).
Projection não é runtime CQRS e não persiste. Repositórios, databases,
brokers, subscriptions, offsets, checkpoints, carga de eventos, processamento
em background e consistência eventual ficam fora de modeling.

## Valores grandes e `hugeParam`

A regra de lint `hugeParam` (e análogas) é um **guarda-corpo arquitetural**
sobre o tamanho de valores concretos que a aplicação escolhe copiar por
assinatura — não um motivo para ponteirizar o metamodelo. `Reducer`,
`Projection`, `Change` e demais tipos de capability permanecem com semântica
de valor (`func` / valores). Estados ou projeções grandes devem ser modelados
na aplicação; não se autoriza `//nolint` nem `Ref`/`Box`/`Readonly` no
contrato de modeling para contornar o lint.

## Fluxo explícito de domínio

```text
Current State + Input
          ↓
Requirement[S,I] ── failure → application error
          ↓ valid
Transition[S,I,O] or captured Change[S]
          ↓
Invariant[O] ────── failure → application error
          ↓ valid
Decision[O,E]
```

A aplicação invoca cada passo explicitamente. modeling não tem pipeline,
runner, command handler, rules engine, registry, dispatcher, state manager,
persistência nem transporte.

AllRules e AllRequirements preservam ordem de declaração, param na primeira
falha, preservam sua identidade e aceitam zero operandos. AnyRule e
AnyRequirement exigem um primeiro operando, param no primeiro sucesso e
devolvem o último erro avaliado quando todos falham. Constituintes nil panican
na construção da composição, alinhados à política de erro de programador de
Invariant.

Change usa contrato separado: Identity representa ausência de mudança; Compose
roda right-to-left; Pipe e ApplyAll rodam left-to-right; composição vazia
devolve Identity; ApplyAll vazio devolve seu input. Change nil tem comportamento
ordinário de função nil em Go, não validação eager de Rule/Requirement.

`MapState` altera somente o estado de Decision; `MapEvents` altera somente os
eventos em ordem. `Emit` e `AppendEvents` preservam o snapshot de entrada.
`Combine` seleciona o estado direito e concatena eventos left-before-right.
FlatMap, negação, bridges de predicado, acumulação de rules, changes
condicionais/guarded, hooks de validação e aliases ficam fora do escopo de
Change.

## Imutabilidade e custo

A Decision pública da raiz é uma fachada sobre Value interno privado apoiado
por lista imutável. Eventos permanecem expostos apenas como `iter.Seq[E]`, o
que mantém modeling evolvível de forma independente. Imutabilidade de Decision
é superficial: estado ou valores de evento referenciados são responsabilidade
do chamador.

Change também é conceitual e superficialmente imutável. Structs podem ser
copiados antes da atualização; slices, maps, ponteiros e referências
compartilhadas permanecem responsabilidade do autor do Change. Não há
reflection nem deep copy. Composição é passagem linear sobre as funções
fornecidas e não define teto de performance independente de máquina.

## Não-objetivos

O módulo não tem reflection, annotations, scanner, registry, service locator,
container de DI, dispatch mágico, command bus, event bus, event store, worker
ou runtime de projection, repositório, ORM, database, HTTP, broker,
filesystem nem tipo concreto de domínio de aplicação.

Tipos Entity/Aggregate/Handle de biblioteca permanecem fora; ver adr:workflow:012 e functional/docs/patterns/cow-handles.md para o padrão de handle na aplicação.

Change não é Patch, Diff, state manager, implementação de Reducer, engine de
Replay, substituto de Transition, produtor de eventos, validador, guard,
runtime, builder nem API fluente.

`functional/state` e `functional/immutable` são capabilities opcionais para
aplicações com estado maior. Não são superclasses, e seus engines internos
nunca pertencem a modeling. `functional/primitives/function.Transform` supre
os algoritmos privados de composição de Change, mas nunca é aliasado,
reexportado ou exposto pela capability de modeling. Reducer e projection, por
sua vez, não dependem de `functional` em produção.
