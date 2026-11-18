---
type: adr
id: adr:workflow:014
context: workflow
title: Evidência AST para o gate de superfície exercitada
status: Aceita
tags: [workflow, archtest, ast, surface, consumers, gates]
deciders: [alvadorn]
related: [adr:workflow:001, adr:workflow:009, adr:workflow:011, adr:workflow:013, tdd:workflow:core]
date: 2026-10-23
---
# ADR-014: Evidência AST para o gate de superfície exercitada

## Contexto e problema

`TestSurfaceIsExercised` existe para impedir que um export entre na superfície
só por simetria: cada símbolo precisa de um consumidor fora do próprio pacote
(`tdd:workflow:core`). O gate original concatenava o texto de todos os arquivos
consumidores e procurava `pacote.Nome` com `strings.Contains`. Texto cru não
distingue uso de menção, e o gate aceitava evidência que não é consumo:

- **prefixo:** `workflow.MapFallible` contém `workflow.Map`, então `Map`
  parecia exercitado sem exemplo próprio; a falta de um `ExampleMap` real ficou
  escondida por isso;
- **comentário e string:** `// action.Tap` ou `"action.Tap"` contavam como uso;
- **pacote estrangeiro:** `outro/action.Tap` ou um identificador `xaction.Tap`
  contêm a mesma substring;
- **identificador sombreado:** uma variável local chamada `action` com um campo
  `Tap` contava como o pacote;
- **falso negativo:** um import com alias (`act "…/workflow/action"`) usa o
  símbolo e não casava com o texto `action.Tap`.

O critério de simplicidade só vale se a evidência for um uso real do símbolo
exportado daquele import path.

## Drivers da decisão

- a evidência precisa ser identificada por import path exato e símbolo exato,
  não por texto;
- sem novo pacote, sem dependência e sem mudar a superfície pública do módulo
  (`go.mod` continua sem `require`);
- o conjunto legítimo de consumidores não pode encolher por acidente;
- tipos carrier retornados por uma função consumida (`Pair`, `AttemptResult`,
  `RetryPolicy.Backoff`) não devem exigir um uso textual separado, mas a regra
  de alcance não pode virar um jeito de creditar qualquer export;
- o gate continua sendo um gate: precisa de evidência negativa de que falha.

## Decisão

O gate lê os consumidores com `go/parser` (`parser.SkipObjectResolution`) e
conta um símbolo somente quando existe um `SelectorExpr` cujo qualificador é um
identificador ligado, por import, ao pacote rastreado, e cujo `Sel` é exatamente
o nome exportado. O gate rastreia os cinco pacotes de produção do módulo
(`workflow`, `action`, `plan`, `resilience`, `transaction`). As chaves de
evidência são `importpath + símbolo`.

### Ligação do qualificador

- o nome local vem do import: alias explícito quando existe; sem alias, o nome
  de pacote conhecido do pacote rastreado;
- import em branco (`_`) é ignorado; import com ponto (`.`) de um pacote
  rastreado faz o gate falhar de forma explícita, porque um uso sem qualificador
  não é verificável por seletor;
- um nome do qualificador só conta fora dos trechos em que um identificador
  local o sombreia. O gate calcula esses trechos lexicais apenas para nomes que
  colidem com os aliases rastreados, seguindo as regras de escopo de Go:
  `var`/`const` valem a partir do fim de cada `ValueSpec`; `:=` vale a partir do
  fim da atribuição até o fim do bloco; tipo local vale do nome do `TypeSpec`;
  parâmetros de tipo valem na assinatura inteira; receiver, parâmetros e
  resultados nomeados valem no corpo de `FuncDecl` e `FuncLit`, não na
  assinatura; `range :=` vale só no corpo; o guard `:=` de `type switch` vale por
  cláusula, e o `:=` de recebimento em `select` vale até o fim da sua cláusula.
  Campo de struct, rótulo, chave de literal composto e nome de parâmetro de um
  `FuncType` autônomo não ligam nada;
- usar o `Obj`/`Scope` produzido pelo parser foi descartado como mecanismo:
  `ast.Object` é marcado como obsoleto na biblioteca padrão. Esta decisão não
  afirma que o linter o rejeitava; ela escolhe o caminho não obsoleto
  independentemente disso.

### Alcance dos carriers

Evidência direta continua sendo o seletor exato. A regra de alcance é estreita
e local à assinatura:

- a partir de uma **função exportada realmente consumida**, seguem-se os tipos
  locais nomeados que aparecem em seus **resultados**, inclusive aninhados
  (`Action[Pair[A, B]]`, `Action[AttemptResult[T]]`);
- a partir de um **tipo consumido**, seguem-se o tipo subjacente e os **tipos**
  dos campos exportados, recursivamente (`RetryPolicy` alcança `Backoff`);
- nomes de campo, parâmetros de função, campos e métodos não exportados,
  identificadores no corpo e declarações não consumidas não contam;
- uma ligação de parâmetro de tipo não é um tipo do pacote que tenha o mesmo
  nome;
- resultado qualificado de outro pacote não é creditado.

Com isso uma função não consumida continua não consumida: `Pair` só é alcançado
via `Zip`, `AttemptResult` só via `Attempt`, `Backoff` só via `RetryPolicy`.
Constantes exigem seletor direto. Os carriers continuam públicos e a superfície
não muda.

### Conjunto de consumidores

O gate lê estes consumidores como AST:

- arquivos `.go` de produção deste módulo — um pacote irmão que usa o símbolo
  conta, como já documenta `TestSurfaceIsExercised`;
- os `example_test.go` externos do próprio módulo. Eles são exemplos compilados
  e executáveis em `package x_test`, ou seja, uso real feito pela API pública, e
  por isso são aceitos.

Ficam de fora `internal/archtest` e os testes unitários comuns do próprio
módulo: um símbolo usado só pelos testes escritos para ele não tem consumidor e
contar esses testes tornaria o gate vacuoso. O filtro usa caminhos relativos
ao módulo, e `go vet ./...` também verifica se os consumidores compilam.

## Alternativas consideradas

- **Manter o `strings.Contains` cru.** Rejeitada: ela é a causa dos falsos
  positivos acima e já aprovou, na prática, um export sem exemplo.
- **Carregar tipos (`go/types` ou `golang.org/x/tools/go/packages`).** Daria
  resolução exata de escopo, alias e dot import. Rejeitada: traria dependência
  externa ou um loader de pacotes dentro de um gate, com custo e fragilidade de
  ambiente, e quebraria a decisão de que o módulo não declara `require`. A
  informação adicional não muda o resultado de um gate de consumo.
- **AST sem tipos, com escopo lexical limitado (escolhida).** Preserva import
  path exato, alias e sombreamento sem loader nem dependência, no estilo dos
  parsers que `support_test.go` e `context_test.go` já usam.
- **Rejeitar os `example_test.go` do módulo.** Rejeitada: os exemplos
  compilados são documentação executável. O gate exige seletor exato e
  rejeita comentários e strings como evidência.
- **Exigir um `var _ = pacote.Nome` dentro de `internal/archtest`.** Rejeitada:
  o próprio gate seria o consumidor, exatamente o vacuous pass que a exclusão de
  `internal/archtest` evita.

## Consequências

`Map` deixa de ser creditado por `MapFallible`, e um uso em comentário, string,
pacote estrangeiro ou identificador sombreado deixa de contar. Import com alias
passa a contar. Os gates ganham casos negativos determinísticos: comentário,
string, prefixo, pacote estrangeiro, import em branco, dot import, alias,
genéricos, sombreamento (antes, no próprio inicializador, depois, restaurado ao
sair do bloco, `var` agrupado, parâmetros contra seletor válido na assinatura,
`FuncLit` aninhada, `if`/`for`/`switch`/`range`/`select`/`type switch`,
parâmetros de tipo, tipo local, rótulo e nome de campo) e a regra de alcance dos
carriers.

**Custos aceitos.**

- A análise é sintática e sem tipos. O gate prova que existe um seletor para o
  import path e o símbolo corretos, não que o uso tem tipos válidos nem que o
  código é executado: selecionar um símbolo não significa execução dinâmica.
  `go vet ./...` verifica os tipos, e `go test ./...` executa os exemplos
  com saída declarada.
- O cálculo de escopo lexical é código próprio que aproxima as regras de Go
  apenas para nomes que colidem com os aliases rastreados. Não há tabela de
  símbolos geral, loader nem framework; redeclaração de um nome já local é
  tratada por união de intervalos sem checagem de tipos, e uma colisão entre
  nome de import e nome de pacote falha na compilação separada, não no gate.
- A regra de alcance cobre só a assinatura local (resultado de função, tipo e
  campos exportados); não é prova geral de que todo tipo é exercitado.
- O conjunto de consumidores aceita os `example_test.go` do próprio módulo e
  pacotes irmãos de produção: o autor do símbolo pode escrever o consumidor. A
  barreira é exigir um uso real, não comentário nem texto.

Nenhuma superfície pública é removida ou criada: `AcquireUseRelease`, `Attempt`
e `AttemptResult` permanecem, e `Retry` não usa `errors.Join`. A mudança toca só
`internal/archtest`, o exemplo compilado `ExampleMap` e a documentação.

## Não-objetivos

Carregar tipos ou dependências externas, ligação por dot import de pacote
rastreado, tabela de símbolos geral, nova proibição de consumidor, allowlist
fixa de carriers, prova de execução dinâmica e qualquer mudança em `action`,
`plan`, `resilience`, `transaction` ou no pacote raiz.
