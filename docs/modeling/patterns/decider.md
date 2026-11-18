# Decider

A modeling pattern, not a library type. A Decider describes an aggregate with
three pieces: `decide` (command and state become events), `evolve` (state and
event become the next state), and the initial state. modeling already has a
shape for each piece, and the application composes them with ordinary
functions.

All code on this page is an excerpt from
[`decider_example_test.go`](../../../modeling/decider_example_test.go), which `go test`
compiles and runs in the module. The law that ties `decide` to `evolve` is in
the [decide/evolve law](decide-evolve-law.md).

## Mapping

| Decider | modeling | In the example |
|---|---|---|
| `decide: C → S → [E]` | `modeling.FallibleTransition[S, C, modeling.Decision[S, E]]`, or `Transition` when the command cannot fail | `decideDeposit`, `decidePoints`, `decideWallet` |
| `evolve: S → E → S` | `reducer.Reducer[S, E]` | `evolveAccount`, `evolvePoints`, `evolveWallet` |
| `initialState: S` | application value | `deciderAccountInitial`, `deciderPointsInitial`, `deciderWalletInitial` |
| history fold | `reducer.Replay(initial, events, evolve)` | `reducer.Replay` inside `decidePoints` |
| rejected command | the application's `error` | `errNonPositiveDeposit`, `errNonPositivePoints` |

The canonical `decide` returns only events. In modeling it returns a
`Decision[S, E]`, which carries the next state along with the events, and
rejection goes in the `error`. The shapes are function types, so ordinary
application functions fit without an adapter:

```go
var _ modeling.FallibleTransition[account, deposit, modeling.Decision[account, accountDeposited]] = decideDeposit

var _ reducer.Reducer[account, accountDeposited] = evolveAccount
```

## Two decide styles

**State-first.** `decideDeposit` computes the next state itself and emits the
event `evolveAccount` must replay to reach the same state. The test of the
[law](decide-evolve-law.md) guarantees the two routes agree:

```go
func decideDeposit(state account, command deposit) (modeling.Decision[account, accountDeposited], error) {
	if command.amount <= 0 {
		return modeling.Decide[accountDeposited](state), errNonPositiveDeposit
	}
	next := account{balance: state.balance + command.amount, deposits: state.deposits + 1}
	return modeling.Decide[accountDeposited](next).Emit(accountDeposited(command)), nil
}
```

**Events-first.** `decidePoints` decides only the events and derives the next
state by replaying them with `evolvePoints`. It is the canonical Decider
`decide`, and the law holds by construction under any observation:

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

Prefer events-first when the events carry the whole change: the Decision state
comes out of `evolve` and cannot diverge. Use state-first when the next state
is clearer computed directly, or when part of it is in no event at all, like
the checkout coupon in
[Partial observation](decide-evolve-law.md#observação-parcial). Then the law
needs a test, with the observation declared. In both styles the Decision
already carries the finished state, and whoever confirms needs no Replay.

<a id="comandos-sem-tipos-soma"></a>

## Commands without sum types

Go has no sum type. The canonical Decider takes a `C` that is the sum of the
aggregate's commands and pattern-matches inside `decide`.

**One `decide` per command.** `decideDeposit` takes `deposit` and `decidePoints`
takes `earnPoints`. Each is a `FallibleTransition` with its own command type,
and the application's entry point (handler, use case) calls the right one.
There is no `Command` interface, dispatcher, registry, or command bus.

**Events use a sealed interface.** The history mixes event types in a single
`iter.Seq[E]`, and `evolve` must accept them all. An interface with an
unexported method plays the role of the sum:

```go
type walletEvent interface {
	walletEvent()
}

func (accountDeposited) walletEvent() {}

func (pointsEarned) walletEvent() {}
```

`evolve` dispatches with a type switch:

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

The compiler does not check that the switch is exhaustive. An event without a
`case` leaves the switch without changing state; if the `decide` emits that
event, the law breaks and the table or the fuzz flags it.

**Sum-typed command, when needed.** A single entry point uses the same shape: a
sealed interface and a type switch in `decide`. `topUp`, by contrast, is a
product command that carries both subcommands at once:

```go
type topUp struct {
	deposit deposit
	earn    earnPoints
}
```

## Composition over a product state

The wallet is the product of the account and the points:

```go
type wallet struct {
	account account
	points  points
}

var deciderWalletInitial = wallet{account: deciderAccountInitial, points: deciderPointsInitial}
```

`decideWallet` decides each component, lifts the result into the wallet with
`MapState` (state) and `MapEvents` (events), and joins the two sides with
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

- **Sequential composition.** `Combine(left, right)` keeps `right`'s state and
  puts `left`'s events before `right`'s. Because `right` was decided over
  `left.State()`, this is sequential composition: first the account, then the
  points over the resulting state. The product law follows from the component
  laws and Replay concatenation, as long as `evolveWallet` forwards each event
  to its component:
  `Replay(before, L ++ R) = Replay(Replay(before, L), R) = Replay(left.State(), R) = right.State()`.
- **Disjoint components.** `decidePoints` only reads `points`, and
  `left.State().points` equals `state.points`. With disjoint components,
  deciding over `left.State()` gives the same result as deciding over the
  previous state. When one component reads what the other changes (points
  proportional to the balance, for example), deciding over the previous state
  would use stale data, and only the sequential form gives "one after the
  other". When lifting `right`, `left`'s state must enter the product:
  `Combine` keeps only `right`'s state, and a `right` built over the previous
  state would discard `left`'s change. The law flags it, because `left`'s
  events are still in the history.
- **Atomic rejection.** If a component rejects, `decideWallet` returns the
  error and a Decision without events over the original state. There is no
  half decision.
- **Canonical product.** In the canonical Decider, the product takes a sum
  command `C1 | C2` and forwards it to a single component. The other component
  stays as it is, and there is no `Combine`. In Go this is the type switch of
  [Commands without sum types](#comandos-sem-tipos-soma) calling a single
  `decide`, lifted with `MapState` and `MapEvents`. `decideWallet` is another
  composition: the product command `topUp` drives both components in sequence.

The executable example shows the event order and the law on the product:

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

`TestDecideEvolveLawTable` and `FuzzDecideEvolveLaw` cover the wallet with the
same recipes as the account.

## What it is not

- **Not a library type.** modeling exports no `Decider`, `Compose`, `Run`,
  runner, or decider registry.
  [minimal type safe domain modeling](../../adr/modeling/minimal-type-safe-domain-modeling.md)
  (`adr:modeling:minimal-type-safe-domain-modeling`) keeps the core minimal and requires a separately
  approved change for any new framework type.
  [The workflow decision on DDD runtimes](../../../workflow/docs/adr/workflow/no-ddd-runtime-module.md)
  (`adr:workflow:no-ddd-runtime-module`) decides that helping DDD applications is pattern
  documentation, not a runtime. A Decider helper in a library would need its
  own ADR.
- **Not Event Sourcing.** Nothing here loads streams, persists, versions, or
  takes snapshots (see the [overview](../overview.md)).
- **Not a command bus.** Choosing the `decide` for each command is application
  code.

## Relation to load → decide → confirm

The Decider is the pure middle of
[load → decide → confirm](../../../workflow/docs/patterns/load-decide-confirm.md):

- **load** produces the previous state. With a state repository, it is the
  state read. With an event history, it is the Replay of the recorded history
  from `deciderWalletInitial` with `evolveWallet`.
- **decide** is `decideWallet`: a pure function, no effect.
- **confirm** writes `d.State()` and appends `d.Events()` in order to the
  application's outbox (see
  [Outbox and inbox](../../../workflow/docs/patterns/outbox-inbox.md)).

The law is what makes confirming state and events together possible: whoever
rebuilds the state from the published events arrives at the same `d.State()`
that was written, within the declared observation.
