package reducer_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/modeling/reducer"
)

type accountEventKind uint8

const (
	eventAccountOpened accountEventKind = iota + 1
	eventMoneyDeposited
	eventAccountClosed
)

type accountState struct {
	balance int
	closed  bool
}

type accountEvent struct {
	kind   accountEventKind
	amount int
}

func reduceAccount(state accountState, event accountEvent) accountState {
	switch event.kind {
	case eventAccountOpened:
		state.balance = event.amount
	case eventMoneyDeposited:
		state.balance += event.amount
	case eventAccountClosed:
		state.closed = true
	}
	return state
}

func ExampleReplay() {
	events := seqOf(
		accountEvent{kind: eventAccountOpened, amount: 100},
		accountEvent{kind: eventMoneyDeposited, amount: 25},
		accountEvent{kind: eventAccountClosed},
	)

	state := reducer.Replay(accountState{}, events, reduceAccount)
	fmt.Printf("balance=%d closed=%t\n", state.balance, state.closed)

	// Output:
	// balance=125 closed=true
}
