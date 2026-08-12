---
type: tdd
id: tdd:modeling:projection
context: modeling
title: Modeling Projection and Project
status: Aceito
tags: [modeling, projection, read-model, iter, generics, laws, testing]
related: [adr:modeling:006, adr:modeling:007, adr:modeling:010, tdd:modeling:reducer-replay]
date: 2026-08-12
---
# TDD: Modeling Projection and Project

## Escopo técnico

Descreve o pacote importável `modeling/projection`. Contratos existentes da
raiz, change, policy, reducer, internal, módulo, dependências e release
permanecem separados; projection não os altera.

Projection deriva uma representação de propriedade da aplicação a partir de
fatos. Reducer evolui estado de domínio a partir de fatos. A forma
computacional é a mesma; a responsabilidade semântica não. Nenhum pacote
importa o outro.

O parâmetro de eventos de `Project` é `iter.Seq[E]` (ADR-010). A escolha
anterior `immutable.List[E]` em ADR-007 fica supersedida para esse parâmetro.

## Contratos públicos

O pacote exporta exatamente:

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Funções ordinárias `func(P, E) P` são atribuíveis e invocáveis diretamente.
P e E são `any` irrestrito; a biblioteca não define Event, ReadModel, View nem
marker. Semântica de valor: não há ponteiro no tipo de capability nem
`Ref`/`Box`/`Readonly`.

## Semântica e leis

Projection é a operação incremental:

```text
P x E -> P
```

Project é a aplicação ordenada e repetida (fold de passagem única):

```text
p1 = project(p0, e1)
p2 = project(p1, e2)
...
pn = project(p(n-1), en)
```

Project é síncrono, invoca a Projection uma vez por evento, preserva a ordem
de encontro da sequência e não paraleliza. Entrada vazia devolve o valor
inicial sem invocação. Entrada não vazia com Projection nil tem o panic
ordinário de função nil em Go.

Lei de concatenação (quando a aplicação define `A ++ B` de forma observável):

```text
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

Uso em lote e incremental são equivalentes porque Project apenas repete a
mesma operação. Chamadas repetidas devolvem resultados iguais quando a
Projection é determinística. O pacote não inspeciona relógios, aleatoriedade,
I/O, estado global, mutação ou pureza.

Devolver o P corrente significa que o evento não afeta a representação.
Política de evento desconhecido pertence à união/switch da aplicação.

Como a sequência é de passagem única, multi-pass exige materialização
explícita **antes** de Project.

## Ponte com Decision

Quando os eventos de uma Decision alimentam um read model:

```go
view := projection.Project(empty, decisionValue.Events(), projectFn)
```

Não é necessário `immutable.CollectList`. A biblioteca não oferece
`ProjectDecision` nem compara estado de Decision com o valor projetado.

## Fronteira de implementação e dependências

Project itera `for event := range events` e encadeia um valor local projetado.
Não chama `Items()`, não importa fold, não materializa slice e não delega a
reducer. Complexidade `O(n * cost(projection))` sem cópia automática de
eventos nem dispatch por reflection.

Imports de produção: somente a biblioteca padrão (`iter`). Testes de
arquitetura devem impor exatamente dois exports, ausência de dependência de
sibling e de `functional/*`, range direto sobre `events` e sem reexport na
raiz.

## Testes e benchmarks

Testes black-box fixam: atribuição de função ordinária; P struct e primitivo;
P com coleção persistente da aplicação; E não comparável; identidade em evento
irrelevante; zero/um/muitos eventos; ordem de encontro; encadeamento do
resultado anterior; equivalência com aplicação sequencial manual;
concatenação; repetição determinística; valor inicial imutável superficial;
equivalência lote/incremental; passagem única (sem segunda varredura implícita);
ponte direta com sequência de Decision quando aplicável.

Benchmarks reportam `ns/op`, `B/op` e `allocs/op` em tamanhos 0, 1, 8, 32,
128, 1024 e 8192 para projeção int, projeção de índice e baseline de loop
direto sobre a sequência. Resultados são evidência, não limiar
independente de máquina. Comparar com `FoldLeft` sobre slice materializado
mediria cópia evitável e fica de fora.

## Projection versus infraestrutura

Projection devolve um valor. A aplicação carrega fatos e views correntes,
invoca Projection ou Project e persiste o resultado na própria camada. Este
pacote não conhece repositório, database, broker, event store, subscription,
offset, checkpoint, consumer group, worker em background, handler
command/query nem runtime de consistência eventual.

## Superfície excluída

Não há FallibleProjection, ContramapEvent, Compose, Zip, Map, ContramapState,
alias Replay, API paralela, marker ReadModel/View/Event, registry, dispatcher,
handler map, reflection, persistência, transporte, serialização, migração de
versão, mudança de dependência fora do stdlib nem alias na raiz.

## Verificação

Executar testes focados de projection e arquitetura (passagem única e stdlib
only), testes do módulo em workspace e `GOWORK=off`, race, vet, tidy-diff,
lint sem caps ocultos, listagem de pacotes, matriz de benchmarks e gates
repositório-wide do plano de implementação quando a entrega incluir código.
