package modeling_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
	"github.com/alvadorncorp/gopackages/modeling/reducer"
)

// The identifiers below back the recipes in docs/patterns/decide-evolve-law.md.

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

// sameAccount is the law's observation. account is comparable and every
// field is carried by an event, so obs is the whole state and == suffices.
func sameAccount(a, b account) bool {
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
	})
}
