package reducer_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/reducer"
)

type reducerAccount struct {
	balance int
}

type moneyDeposited struct {
	amount int
}

func applyDeposit(account reducerAccount, event moneyDeposited) reducerAccount {
	account.balance += event.amount
	return account
}

var _ reducer.Reducer[reducerAccount, moneyDeposited] = applyDeposit

func TestReducerAdaptsAndInvokesOrdinaryFunction(t *testing.T) {
	t.Parallel()

	var reduce reducer.Reducer[reducerAccount, moneyDeposited] = applyDeposit
	initial := reducerAccount{balance: 100}

	got := reduce(initial, moneyDeposited{amount: 25})

	if got.balance != 125 {
		t.Fatalf("reduced balance = %d, want 125", got.balance)
	}
	if initial.balance != 100 {
		t.Fatalf("initial balance = %d, want unchanged 100", initial.balance)
	}
}
