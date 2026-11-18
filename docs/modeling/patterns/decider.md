# Decider

Padrão de modelagem, não tipo de biblioteca. Um Decider descreve um agregado
com três peças: `decide` (comando e estado viram eventos), `evolve` (estado e
evento viram o próximo estado) e o estado inicial. modeling já tem uma forma
para cada peça, e a aplicação as compõe com funções ordinárias.

Todo código desta página é trecho de
[`decider_example_test.go`](../../../modeling/decider_example_test.go), que `go test`
compila e executa no módulo. A lei que liga `decide` a `evolve` está em
[Lei decide/evolve](decide-evolve-law.md).

## Mapeamento

| Decider | modeling | No exemplo |
|---|---|---|
| `decide: C → S → [E]` | `modeling.FallibleTransition[S, C, modeling.Decision[S, E]]`, ou `Transition` quando o comando não falha | `decideDeposit`, `decidePoints`, `decideWallet` |
| `evolve: S → E → S` | `reducer.Reducer[S, E]` | `evolveAccount`, `evolvePoints`, `evolveWallet` |
| `initialState: S` | valor da aplicação | `deciderAccountInitial`, `deciderPointsInitial`, `deciderWalletInitial` |
| fold do histórico | `reducer.Replay(initial, events, evolve)` | `reducer.Replay` dentro de `decidePoints` |
| comando rejeitado | `error` da aplicação | `errNonPositiveDeposit`, `errNonPositivePoints` |

O `decide` canônico devolve só eventos. Em modeling ele devolve uma
`Decision[S, E]`, que leva o próximo estado junto com os eventos, e a rejeição
vai no `error`. As formas são tipos de função, então funções comuns da
aplicação encaixam sem adaptador:

```go
var _ modeling.FallibleTransition[account, deposit, modeling.Decision[account, accountDeposited]] = decideDeposit

var _ reducer.Reducer[account, accountDeposited] = evolveAccount
```

## Dois estilos de decide

**State-first.** `decideDeposit` calcula o próximo estado por conta própria e
emite o evento que `evolveAccount` precisa reaplicar até o mesmo estado. Quem
garante que as duas rotas concordam é o teste da
[lei](decide-evolve-law.md):

```go
func decideDeposit(state account, command deposit) (modeling.Decision[account, accountDeposited], error) {
	if command.amount <= 0 {
		return modeling.Decide[accountDeposited](state), errNonPositiveDeposit
	}
	next := account{balance: state.balance + command.amount, deposits: state.deposits + 1}
	return modeling.Decide[accountDeposited](next).Emit(accountDeposited(command)), nil
}
```

**Events-first.** `decidePoints` decide só os eventos e deriva o próximo estado
reaplicando-os com `evolvePoints`. É o `decide` canônico do Decider, e a lei
vale por construção, qualquer que seja a observação:

```go
func decidePoints(state points, command earnPoints) (modeling.Decision[points, pointsEarned], error) {
	if command.amount <= 0 {
		return modeling.Decide[pointsEarned](state), errNonPositivePoints
	}
	events := []pointsEarned{pointsEarned(command)}
	next := reducer.Replay(state, slices.Values(events), evolvePoints)
	return modeling.Decide[pointsEarned](next).Emit(events...), nil
}

func evolvePoints(state points, event pointsEarned) points {
	state.total += event.amount
	return state
}
```

Prefira events-first quando os eventos carregam toda a mudança: o estado da
Decision sai de `evolve` e não tem como divergir. Use state-first quando o
próximo estado é mais claro calculado direto, ou quando parte dele não está
em evento nenhum, como o cupom do checkout em
[Observação parcial](decide-evolve-law.md#observação-parcial). Nesse caso a
lei precisa de teste, com a observação declarada. Nos dois estilos a Decision
já leva o estado pronto, e quem confirma não precisa de Replay.

## Comandos sem tipos-soma

Go não tem tipo-soma. O Decider canônico recebe um `C` que é a soma dos
comandos do agregado e faz o casamento de padrões dentro do `decide`.

**Um `decide` por comando.** `decideDeposit` recebe `deposit` e
`decidePoints` recebe `earnPoints`. Cada um é uma `FallibleTransition` com o
próprio tipo de comando, e o ponto de entrada da aplicação (handler, caso de
uso) chama o certo. Não há interface `Command`, dispatcher, registry nem
command bus.

**Eventos usam interface selada.** O histórico mistura tipos de evento num só
`iter.Seq[E]`, e `evolve` precisa aceitar todos. Uma interface com método não
exportado faz o papel da soma:

```go
type walletEvent interface {
	walletEvent()
}

func (accountDeposited) walletEvent() {}

func (pointsEarned) walletEvent() {}
```

`evolve` despacha com type switch:

```go
func evolveWallet(state wallet, event walletEvent) wallet {
	switch e := event.(type) {
	case accountDeposited:
		state.account = evolveAccount(state.account, e)
	case pointsEarned:
		state.points = evolvePoints(state.points, e)
	}
	return state
}
```

O compilador não verifica se o switch é exaustivo. Um evento sem `case` sai do
switch sem mudar o estado; se o `decide` emite esse evento, a lei quebra e a
tabela ou o fuzz acusa.

**Comando-soma, quando precisar.** Um ponto de entrada único usa a mesma
forma: interface selada e type switch no `decide`. `topUp`, ao contrário, é
comando-produto e leva os dois subcomandos de uma vez:

```go
type topUp struct {
	deposit deposit
	earn    earnPoints
}
```

## Composição sobre estado-produto

A carteira é o produto da conta e dos pontos:

```go
type wallet struct {
	account account
	points  points
}

var deciderWalletInitial = wallet{account: deciderAccountInitial, points: deciderPointsInitial}
```

`decideWallet` decide cada componente, leva o resultado para a carteira com
`MapState` (estado) e `MapEvents` (eventos) e junta os dois lados com
`Combine`:

```go
func decideWallet(state wallet, command topUp) (modeling.Decision[wallet, walletEvent], error) {
	deposited, err := decideDeposit(state.account, command.deposit)
	if err != nil {
		return modeling.Decide[walletEvent](state), err
	}
	left := modeling.MapEvents(modeling.MapState(deposited, func(next account) wallet {
		return wallet{account: next, points: state.points}
	}), toWalletEvent[accountDeposited])

	earned, err := decidePoints(left.State().points, command.earn)
	if err != nil {
		return modeling.Decide[walletEvent](state), err
	}
	right := modeling.MapEvents(modeling.MapState(earned, func(next points) wallet {
		return wallet{account: left.State().account, points: next}
	}), toWalletEvent[pointsEarned])

	return modeling.Combine(left, right), nil
}

func toWalletEvent[E walletEvent](event E) walletEvent {
	return event
}
```

- **Composição sequencial.** `Combine(left, right)` fica com o estado de
  `right` e põe os eventos de `left` antes dos de `right`. Como `right` foi
  decidido sobre `left.State()`, isso é composição sequencial: primeiro a
  conta, depois os pontos sobre o estado resultante. A lei do produto sai das
  leis dos componentes e da concatenação do Replay, desde que `evolveWallet`
  encaminhe cada evento ao seu componente:
  `Replay(before, L ++ R) = Replay(Replay(before, L), R) = Replay(left.State(), R) = right.State()`.
- **Componentes disjuntos.** `decidePoints` só lê `points`, e
  `left.State().points` é igual a `state.points`. Com componentes disjuntos,
  decidir sobre `left.State()` dá o mesmo que decidir sobre o estado anterior.
  Quando um componente lê o que o outro muda (pontos proporcionais ao saldo,
  por exemplo), decidir sobre o estado anterior usaria dados velhos, e só a
  forma sequencial dá "um depois do outro". Ao levantar `right`, o estado de
  `left` precisa entrar no produto: `Combine` guarda só o estado de `right`, e
  um `right` montado sobre o estado anterior descartaria a mudança de `left`.
  A lei acusa, porque os eventos de `left` continuam no histórico.
- **Rejeição atômica.** Se um componente rejeita, `decideWallet` devolve o
  erro e uma Decision sem eventos sobre o estado original. Não existe meia
  decisão.
- **Produto canônico.** No Decider canônico, o produto recebe um comando-soma
  `C1 | C2` e o encaminha a um único componente. O outro componente fica como
  está e não há `Combine`. Em Go, isso é o type switch de
  [Comandos sem tipos-soma](#comandos-sem-tipos-soma) chamando um único
  `decide`, levantado com `MapState` e `MapEvents`. `decideWallet` é outra
  composição: o comando-produto `topUp` aciona os dois componentes em
  sequência.

O exemplo executável mostra a ordem dos eventos e a lei no produto:

```go
func Example_deciderComposition() {
	before := deciderWalletInitial
	d, err := decideWallet(before, topUp{deposit: deposit{amount: 100}, earn: earnPoints{amount: 10}})
	if err != nil {
		fmt.Println("rejected:", err)
		return
	}
	for event := range d.Events() {
		fmt.Printf("%T %+v\n", event, event)
	}
	fmt.Printf("state=%+v\n", d.State())
	fmt.Printf("law=%t\n", decideEvolveConsistent(before, d, evolveWallet, sameWallet))

	// Output:
	// modeling_test.accountDeposited {amount:100}
	// modeling_test.pointsEarned {amount:10}
	// state={account:{balance:100 deposits:1} points:{total:10}}
	// law=true
}
```

`TestDecideEvolveLawTable` e `FuzzDecideEvolveLaw` cobrem a carteira com as
mesmas receitas da conta.

## Não é

- **Não é tipo de biblioteca.** modeling não exporta `Decider`, `Compose`,
  `Run`, runner nem registry de deciders.
  [minimal type safe domain modeling](../../adr/modeling/minimal-type-safe-domain-modeling.md)
  (`adr:modeling:minimal-type-safe-domain-modeling`) mantém o núcleo mínimo e exige mudança aprovada à parte
  para qualquer tipo novo de framework.
  [A decisão sobre runtime DDD no workflow](../../../workflow/docs/adr/workflow/no-ddd-runtime-module.md)
  (`adr:workflow:no-ddd-runtime-module`) decide que a ajuda a aplicações DDD é documentação de
  padrões, não runtime. Um helper de Decider numa biblioteca exigiria ADR
  própria.
- **Não é Event Sourcing.** Nada aqui carrega streams, persiste, versiona nem
  tira snapshot (ver [overview](../overview.md)).
- **Não é command bus.** Escolher o `decide` de cada comando é código da
  aplicação.

## Relação com load → decide → confirm

O Decider é o meio puro de
[load → decide → confirm](../../../workflow/docs/patterns/load-decide-confirm.md):

- **load** produz o estado anterior. Com repositório de estado, é o estado
  lido. Com histórico de eventos, é o Replay do histórico gravado a partir de
  `deciderWalletInitial` com `evolveWallet`.
- **decide** é `decideWallet`: função pura, sem efeito.
- **confirm** grava `d.State()` e faz append sequencial de `d.Events()` no
  outbox da aplicação (ver
  [Outbox e inbox](../../../workflow/docs/patterns/outbox-inbox.md)).

A lei é o que permite confirmar estado e eventos juntos: quem reconstrói o
estado a partir dos eventos publicados chega ao mesmo `d.State()` que foi
gravado, dentro da observação declarada.
