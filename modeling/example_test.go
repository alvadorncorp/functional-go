package modeling_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alvadorncorp/gopackages/functional/immutable"
	"github.com/alvadorncorp/gopackages/functional/state"
	"github.com/alvadorncorp/gopackages/modeling"
)

type (
	exampleOrderID   string
	exampleProductID string
	exampleQuantity  int
)

type exampleMoney struct {
	cents    int64
	currency string
}

func newExampleMoney(cents int64, currency string) (exampleMoney, error) {
	if cents < 0 || currency == "" {
		return exampleMoney{}, errors.New("invalid money")
	}
	return exampleMoney{cents: cents, currency: currency}, nil
}

type exampleOrderItem struct {
	product  exampleProductID
	quantity exampleQuantity
	price    exampleMoney
}

type exampleOrderState struct {
	id    exampleOrderID
	open  bool
	items immutable.List[exampleOrderItem]
}

type exampleAddItem struct {
	product  exampleProductID
	quantity exampleQuantity
	price    exampleMoney
}

type exampleOrderEvent interface {
	exampleOrderEvent()
}

type exampleOrderItemAdded struct {
	product  exampleProductID
	quantity exampleQuantity
}

func (exampleOrderItemAdded) exampleOrderEvent() {}

var (
	errExampleClosedOrder = errors.New("order is closed")
	errExampleQuantity    = errors.New("quantity must be positive")
)

var exampleOrderInvariant = modeling.FailFast(
	modeling.Invariant[exampleOrderState](func(order exampleOrderState) error {
		for item := range order.items.Values() {
			if item.quantity <= 0 {
				return errExampleQuantity
			}
		}
		return nil
	}),
)

func exampleAddItemTransition(
	current *exampleOrderState,
	command exampleAddItem,
) (modeling.Decision[exampleOrderState, exampleOrderEvent], error) {
	unchanged := modeling.Decide[exampleOrderEvent](*current)
	if !current.open {
		return unchanged, errExampleClosedOrder
	}

	next := *current
	next.items = next.items.Append(exampleOrderItem(command))
	if err := exampleOrderInvariant(next); err != nil {
		return unchanged, err
	}

	return modeling.Decide[exampleOrderEvent](next).Emit(exampleOrderItemAdded{
		product: command.product, quantity: command.quantity,
	}), nil
}

func Example() {
	price, _ := newExampleMoney(2500, "USD")
	current := exampleOrderState{id: "order-1", open: true}
	command := exampleAddItem{product: "book", quantity: 2, price: price}
	var transition modeling.FallibleTransition[
		*exampleOrderState,
		exampleAddItem,
		modeling.Decision[exampleOrderState, exampleOrderEvent],
	] = exampleAddItemTransition

	decision, err := transition.Apply(&current, command)
	fmt.Println(decision.State().id, decision.State().items.Len(), decision.EventCount(), err)
	for event := range decision.Events() {
		if added, ok := event.(exampleOrderItemAdded); ok {
			fmt.Println("added", added.product, added.quantity)
		}
	}

	// Output:
	// order-1 1 1 <nil>
	// added book 2
}

func ExampleDecision_withFunctionalState() {
	technical := state.Of(exampleOrderState{id: "order-2", open: true})
	technical = technical.Apply(func(order exampleOrderState) exampleOrderState {
		order.items = order.items.Append(exampleOrderItem{product: "pen", quantity: 1})
		return order
	})
	decision := modeling.Decide[exampleOrderEvent](technical.Materialize())
	fmt.Println(decision.State().items.Len(), decision.EventCount())

	// Output:
	// 1 0
}

func ExampleDecision_composition() {
	left := modeling.Decide[string](1)
	left = modeling.Emit(left, "opened")
	left = modeling.AppendEvents(left, "item-added", "priced")
	right := modeling.Decide[string](2).Emit("closed")

	combined := modeling.Combine(left, right)
	labeled := modeling.MapState(combined, func(state int) string {
		return fmt.Sprintf("state-%d", state)
	})
	upper := modeling.MapEvents(labeled, strings.ToUpper)

	fmt.Println(upper.State())
	for event := range upper.Events() {
		fmt.Println(event)
	}

	// Output:
	// state-2
	// OPENED
	// ITEM-ADDED
	// PRICED
	// CLOSED
}
