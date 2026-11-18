package change_test

import (
	"fmt"
	"testing"

	"go.alvadorn.org/functional/modeling/change"
)

type order struct {
	status string
}

func markPaid(value order) order {
	value.status = "paid"
	return value
}

type identityFactory[S any] func() change.Change[S]

var (
	_ change.Change[order]   = markPaid
	_ identityFactory[order] = change.Identity[order]
)

func TestChangeAcceptsOrdinaryFunction(t *testing.T) {
	current := order{status: "pending"}
	next := change.Change[order](markPaid)(current)

	if next.status != "paid" {
		t.Fatalf("next status = %q, want paid", next.status)
	}
	if current.status != "pending" {
		t.Fatalf("current status = %q, want pending", current.status)
	}
}

func TestIdentityReturnsInputUnchanged(t *testing.T) {
	current := order{status: "pending"}
	if got := change.Identity[order]()(current); got != current {
		t.Fatalf("Identity()(current) = %+v, want %+v", got, current)
	}
}

func ExampleChange() {
	current := order{status: "pending"}
	next := change.Change[order](markPaid)(current)

	fmt.Println(current.status, next.status)
	// Output: pending paid
}
