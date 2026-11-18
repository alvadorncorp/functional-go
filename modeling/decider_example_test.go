package modeling_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/modeling/reducer"
)

// The identifiers below back the recipes in docs/patterns/decide-evolve-law.md
// and docs/patterns/decider.md.

// account is the state of the toy account decider.
type account struct {
	balance  int
	deposits int
}

// deposit is the account command.
type deposit struct {
	amount int
}

// accountDeposited is the account event. It carries every field that
// decideDeposit changes.
type accountDeposited struct {
	amount int
}

var errNonPositiveDeposit = errors.New("deposit must be positive")

// deciderAccountInitial is the account state before any event.
var deciderAccountInitial = account{}

// decideDeposit is state-first: it computes the next state on its own and
// emits the event that evolveAccount must replay into that same state.
func decideDeposit(state account, command deposit) (modeling.Decision[account, accountDeposited], error) {
	if command.amount <= 0 {
		return modeling.Decide[accountDeposited](state), errNonPositiveDeposit
	}
	next := account{balance: state.balance + command.amount, deposits: state.deposits + 1}
	return modeling.Decide[accountDeposited](next).Emit(accountDeposited(command)), nil
}

// evolveAccount applies one account event.
func evolveAccount(state account, event accountDeposited) account {
	state.balance += event.amount
	state.deposits++
	return state
}

// decideDeposit and evolveAccount are ordinary functions of the modeling shapes.
var _ modeling.FallibleTransition[account, deposit, modeling.Decision[account, accountDeposited]] = decideDeposit

var _ reducer.Reducer[account, accountDeposited] = evolveAccount

// points is the state of the toy points decider.
type points struct {
	total int
}

// earnPoints is the points command.
type earnPoints struct {
	amount int
}

// pointsEarned is the points event.
type pointsEarned struct {
	amount int
}

var errNonPositivePoints = errors.New("points must be positive")

// deciderPointsInitial is the points state before any event.
var deciderPointsInitial = points{}

// decidePoints is events-first: it decides only the events and derives the
// next state by replaying them, so the law holds by construction.
func decidePoints(state points, command earnPoints) (modeling.Decision[points, pointsEarned], error) {
	if command.amount <= 0 {
		return modeling.Decide[pointsEarned](state), errNonPositivePoints
	}
	events := []pointsEarned{pointsEarned(command)}
	next := reducer.Replay(state, slices.Values(events), evolvePoints)
	return modeling.Decide[pointsEarned](next).Emit(events...), nil
}

// evolvePoints applies one points event.
func evolvePoints(state points, event pointsEarned) points {
	state.total += event.amount
	return state
}

// wallet is the product state of the account and points deciders.
type wallet struct {
	account account
	points  points
}

// walletEvent is the wallet event. Go has no sum type, so a sealed interface
// stands for accountDeposited | pointsEarned.
type walletEvent interface {
	walletEvent()
}

func (accountDeposited) walletEvent() {}

func (pointsEarned) walletEvent() {}

// topUp is the wallet command: the product of a deposit and the points it earns.
type topUp struct {
	deposit deposit
	earn    earnPoints
}

// deciderWalletInitial is the wallet state before any event.
var deciderWalletInitial = wallet{account: deciderAccountInitial, points: deciderPointsInitial}

// decideWallet composes the component deciders sequentially: the points
// decision is taken on left.State(), and Combine keeps its state after the
// account events. A rejection by either component rejects the whole command.
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

// toWalletEvent lifts a component event into walletEvent for MapEvents.
func toWalletEvent[E walletEvent](event E) walletEvent {
	return event
}

// evolveWallet routes each event to the evolve of its component.
func evolveWallet(state wallet, event walletEvent) wallet {
	switch e := event.(type) {
	case accountDeposited:
		state.account = evolveAccount(state.account, e)
	case pointsEarned:
		state.points = evolvePoints(state.points, e)
	}
	return state
}

// sameAccount is the law's observation. account is comparable and every
// field is carried by an event, so obs is the whole state and == suffices.
func sameAccount(a, b account) bool {
	return a == b
}

// sameWallet is the wallet observation: the whole comparable product state.
func sameWallet(a, b wallet) bool {
	return a == b
}

// decideEvolveConsistent reports whether replaying the events of d onto before
// observes the same state as d.State(): obs(Replay(before, d.Events(), evolve)) == obs(d.State()).
func decideEvolveConsistent[S, E any](
	before S,
	d modeling.Decision[S, E],
	evolve reducer.Reducer[S, E],
	equal func(S, S) bool,
) bool {
	return equal(reducer.Replay(before, d.Events(), evolve), d.State())
}

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

// decideDepositWithoutEvent is a broken decider: it changes the state but emits
// no event, so evolve has nothing to replay.
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
