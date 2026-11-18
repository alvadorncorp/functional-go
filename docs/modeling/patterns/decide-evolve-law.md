# Lei decide/evolve

Padrão de modelagem, não API. Um decider state-first calcula o próximo estado
por conta própria (`Decision.State()`) e emite os fatos (`Decision.Events()`)
que um `reducer.Reducer` reaplica. As duas rotas precisam concordar sobre
aquilo que os eventos prometem reconstruir.

Todo código desta página é trecho de
[`decider_example_test.go`](../../../modeling/decider_example_test.go), que `go test`
compila e executa no módulo.

## Enunciado

Para toda decisão bem-sucedida `d` tomada a partir de `before`:

```text
obs(Replay(before, d.Events(), evolve)) == obs(d.State())
```

- `obs` é a observação escolhida pela aplicação: a parte do estado que os
  eventos se comprometem a reconstruir. `obs` igual à identidade é o caso mais
  forte. Uma observação parcial declara o que os eventos não carregam (ver
  [Observação parcial](#observação-parcial)).
- `==` só serve quando a observação é comparável em Go, como a struct
  `account` do exemplo. Estado com `immutable.List`, `Map` ou `Set`, dinheiro
  com moeda ou outro value object compara pelo `Equal`/`EqualFunc` do próprio
  tipo, campo a campo. `reflect.DeepEqual` não serve: compara representação
  interna, não valor.
- Decisões rejeitadas (erro, ou `Result` em `Err`) ficam fora da lei, porque
  não há Decision a confrontar.

O predicado recebe a observação embutida em `equal(a, b)`, que significa
`obs(a) == obs(b)`:

```go
func decideEvolveConsistent[S, E any](
	before S,
	d modeling.Decision[S, E],
	evolve reducer.Reducer[S, E],
	equal func(S, S) bool,
) bool {
	return equal(reducer.Replay(before, d.Events(), evolve), d.State())
}
```

No exemplo, `decideDeposit` é o decide state-first, `evolveAccount` é o evolve
e `deciderAccountInitial` é o estado inicial. Os dois são funções ordinárias
das formas de modeling:

```go
var _ modeling.FallibleTransition[account, deposit, modeling.Decision[account, accountDeposited]] = decideDeposit

var _ reducer.Reducer[account, accountDeposited] = evolveAccount
```

Como `account` é comparável e `accountDeposited` carrega tudo o que
`decideDeposit` muda, a observação é o estado inteiro:

```go
func sameAccount(a, b account) bool {
	return a == b
}
```

## Lei da aplicação, não da biblioteca

modeling não compara `Decision.State()` com o resultado de Replay e não
oferece `VerifyDecision` nem `ReplayDecision`
([TDD Reducer e Replay, linhas 95-98](../../tdd/modeling/reducer-replay.md)).
Decision e Reducer continuam ortogonais. Eventos de integração não são
universalmente reconstrutores de estado, e só a aplicação sabe qual `obs` é
honesta para o seu agregado.

A lei vive nos testes da aplicação. O predicado acima tem uma linha de corpo:
copie-o para o `_test.go` do domínio em vez de esperar um helper exportado.

## Corolários

**1. Mudança observável exige evento.** Como `Replay(s, vazio, r) = s`, uma
decisão sem eventos só satisfaz a lei se `obs(d.State()) == obs(before)`. Um
decider que altera o estado em silêncio quebra a lei, e o predicado detecta:

```go
func decideDepositWithoutEvent(state account, command deposit) modeling.Decision[account, accountDeposited] {
	state.balance += command.amount
	state.deposits++
	return modeling.Decide[accountDeposited](state)
}

func TestDecideEvolveLawRejectsStateChangeWithoutEvent(t *testing.T) {
	before := account{balance: 10, deposits: 1}
	d := decideDepositWithoutEvent(before, deposit{amount: 5})

	if decideEvolveConsistent(before, d, evolveAccount, sameAccount) {
		t.Fatalf("law held for a silent change: before = %+v, State() = %+v", before, d.State())
	}
}
```

O mesmo vale campo a campo: quando a decisão emite eventos, eles precisam
carregar toda mudança observável que ela fez.

**2. Histórico.** Se cada passo satisfaz a lei, reaplicar todos os eventos
emitidos a partir do estado inicial reproduz o estado final. Isso segue da lei
de concatenação `Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)`. Com
observação parcial, o corolário exige também que `evolve` não leia o que `obs`
descarta. Sem isso, dois estados com a mesma observação podem divergir no
passo seguinte.

## Receita de tabela

Um caso por comando relevante. Cada caso decide a partir de `before` e
confronta a Decision com o Replay dos próprios eventos. A função tem uma
tabela por decider; a segunda cobre a carteira, o produto descrito em
[Decider](decider.md):

```go
func TestDecideEvolveLawTable(t *testing.T) {
	tests := []struct {
		name    string
		before  account
		command deposit
	}{
		{name: "first deposit", before: deciderAccountInitial, command: deposit{amount: 10}},
		{name: "later deposit", before: account{balance: 10, deposits: 1}, command: deposit{amount: 5}},
		{name: "large deposit", before: account{balance: 15, deposits: 2}, command: deposit{amount: 1_000_000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decideDeposit(tt.before, tt.command)
			if err != nil {
				t.Fatalf("decideDeposit(%+v, %+v) error = %v", tt.before, tt.command, err)
			}
			if !decideEvolveConsistent(tt.before, d, evolveAccount, sameAccount) {
				replayed := reducer.Replay(tt.before, d.Events(), evolveAccount)
				t.Fatalf("Replay = %+v, State() = %+v", replayed, d.State())
			}
		})
	}

	walletTests := []struct {
		name    string
		before  wallet
		command topUp
	}{
		{
			name:    "first top-up",
			before:  deciderWalletInitial,
			command: topUp{deposit: deposit{amount: 10}, earn: earnPoints{amount: 1}},
		},
		{
			name:    "later top-up",
			before:  wallet{account: account{balance: 10, deposits: 1}, points: points{total: 1}},
			command: topUp{deposit: deposit{amount: 5}, earn: earnPoints{amount: 3}},
		},
	}
	for _, tt := range walletTests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decideWallet(tt.before, tt.command)
			if err != nil {
				t.Fatalf("decideWallet(%+v, %+v) error = %v", tt.before, tt.command, err)
			}
			if !decideEvolveConsistent(tt.before, d, evolveWallet, sameWallet) {
				replayed := reducer.Replay(tt.before, d.Events(), evolveWallet)
				t.Fatalf("Replay = %+v, State() = %+v", replayed, d.State())
			}
		})
	}
}
```

A tabela é o lugar dos casos de borda conhecidos do domínio, como
arredondamento e valor zero.

## Receita de fuzz

O fuzz gera uma sequência de comandos, verifica a lei em cada passo e, no fim,
o corolário de histórico. A mesma entrada dirige a conta e a carteira; na
carteira, o valor 1 rende zero pontos e exercita a rejeição depois que o
depósito já decidiu:

```go
func FuzzDecideEvolveLaw(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{10})
	f.Add([]byte{10, 0, 255, 1})
	f.Fuzz(func(t *testing.T, amounts []byte) {
		state := deciderAccountInitial
		var history []accountDeposited
		for _, amount := range amounts {
			d, err := decideDeposit(state, deposit{amount: int(amount)})
			if err != nil {
				continue // a rejected command decides nothing
			}
			if !decideEvolveConsistent(state, d, evolveAccount, sameAccount) {
				t.Fatalf("step law: before = %+v, State() = %+v", state, d.State())
			}
			state = d.State()
			history = slices.AppendSeq(history, d.Events())
		}

		replayed := reducer.Replay(deciderAccountInitial, slices.Values(history), evolveAccount)
		if !sameAccount(replayed, state) {
			t.Fatalf("history corollary: Replay = %+v, final State() = %+v", replayed, state)
		}

		walletState := deciderWalletInitial
		var walletHistory []walletEvent
		for _, amount := range amounts {
			// amount 1 earns 0 points: the deposit decides, then decidePoints rejects the top-up
			command := topUp{deposit: deposit{amount: int(amount)}, earn: earnPoints{amount: int(amount) / 2}}
			d, err := decideWallet(walletState, command)
			if err != nil {
				continue
			}
			if !decideEvolveConsistent(walletState, d, evolveWallet, sameWallet) {
				t.Fatalf("wallet step law: before = %+v, State() = %+v", walletState, d.State())
			}
			walletState = d.State()
			walletHistory = slices.AppendSeq(walletHistory, d.Events())
		}

		walletReplayed := reducer.Replay(deciderWalletInitial, slices.Values(walletHistory), evolveWallet)
		if !sameWallet(walletReplayed, walletState) {
			t.Fatalf("wallet history corollary: Replay = %+v, final State() = %+v", walletReplayed, walletState)
		}
	})
}
```

`go test` roda só as sementes. Para explorar localmente:

```bash
cd modeling && go test -run '^$' -fuzz FuzzDecideEvolveLaw -fuzztime 10s .
```

Uma falha grava o caso em `testdata/fuzz/` dentro do pacote. Transforme o caso
em linha da tabela ou em semente `f.Add` e não versione o corpus.

## Observação parcial

Por exemplo, num catálogo, o replay pode reconstruir título e versão a
partir dos eventos, mas não uma nota de edição que nunca foi registrada em
evento. A observação `obs` deve excluir a nota e explicar o motivo.

Se o campo faz parte do contrato de replay, excluí-lo da comparação esconderia
um defeito: nesse caso, escreva o teste com o campo e corrija a decisão ou o
evento que deveria carregar a informação.
