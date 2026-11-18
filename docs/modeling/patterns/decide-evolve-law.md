# Decide/evolve law

A modeling pattern, not an API. A state-first decider computes the next state
itself (`Decision.State()`) and emits the facts (`Decision.Events()`) that a
`reducer.Reducer` replays. The two routes must agree on what the events promise
to rebuild.

All code on this page is an excerpt from
[`decider_example_test.go`](../../../modeling/decider_example_test.go), which `go test`
compiles and runs in the module.

## Statement

For every successful decision `d` made from `before`:

```text
obs(Replay(before, d.Events(), evolve)) == obs(d.State())
```

- `obs` is the observation the application chooses: the part of the state the
  events commit to rebuilding. `obs` equal to the identity is the strongest
  case. A partial observation declares what the events do not carry (see
  [Partial observation](#observação-parcial)).
- `==` only works when the observation is comparable in Go, like the example's
  `account` struct. State with an `immutable.List`, `Map`, or `Set`, money with
  a currency, or another value object compares through that type's own
  `Equal`/`EqualFunc`, field by field. `reflect.DeepEqual` does not work: it
  compares internal representation, not value.
- Rejected decisions (an error, or a `Result` in `Err`) stay outside the law,
  because there is no Decision to confront.

The predicate receives the observation embedded in `equal(a, b)`, which means
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

In the example, `decideDeposit` is the state-first decide, `evolveAccount` is
the evolve, and `deciderAccountInitial` is the initial state. Both are ordinary
functions of the modeling shapes:

```go
var _ modeling.FallibleTransition[account, deposit, modeling.Decision[account, accountDeposited]] = decideDeposit

var _ reducer.Reducer[account, accountDeposited] = evolveAccount
```

Because `account` is comparable and `accountDeposited` carries everything
`decideDeposit` changes, the observation is the whole state:

```go
func sameAccount(a, b account) bool {
	return a == b
}
```

## An application law, not a library law

modeling does not compare `Decision.State()` with the Replay result and offers
neither `VerifyDecision` nor `ReplayDecision`
([TDD Reducer and Replay, lines 97-99](../../tdd/modeling/reducer-replay.md)).
Decision and Reducer stay orthogonal. Integration events are not universal
state rebuilders, and only the application knows which `obs` is honest for its
aggregate.

The law lives in the application's tests. The predicate above has a one-line
body: copy it into the domain's `_test.go` instead of waiting for an exported
helper.

## Corollaries

**1. Observable change requires an event.** Since `Replay(s, empty, r) = s`, a
decision without events satisfies the law only if
`obs(d.State()) == obs(before)`. A decider that changes state silently breaks
the law, and the predicate catches it:

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

The same holds field by field: when the decision emits events, they must carry
every observable change it made.

**2. History.** If every step satisfies the law, replaying all emitted events
from the initial state reproduces the final state. This follows from the
concatenation law `Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)`. With
a partial observation, the corollary also requires that `evolve` never reads
what `obs` discards. Without that, two states with the same observation can
diverge on the next step.

## Table recipe

One case per relevant command. Each case decides from `before` and confronts
the Decision with the Replay of its own events. The test has one table per
decider; the second covers the wallet, the product described in
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

The table is where the domain's known edge cases live, such as rounding and
zero values.

## Fuzz recipe

The fuzz generates a sequence of commands, checks the law at each step, and
checks the history corollary at the end. The same input drives the account and
the wallet; in the wallet, the value 1 earns zero points and exercises
rejection after the deposit has already decided:

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

`go test` runs only the seeds. To explore locally:

```bash
cd modeling && go test -run '^$' -fuzz FuzzDecideEvolveLaw -fuzztime 10s .
```

A failure writes the case to `testdata/fuzz/` inside the package. Turn the case
into a table row or an `f.Add` seed, and do not version the corpus.

<a id="observação-parcial"></a>

## Partial observation

For example, in a catalog, the replay can rebuild the title and version from
the events but not an editing note that was never recorded as an event. The
observation `obs` must exclude the note and explain why.

If the field is part of the replay contract, excluding it from the comparison
would hide a defect: in that case, write the test with the field and fix the
decision or the event that should have carried the information.
