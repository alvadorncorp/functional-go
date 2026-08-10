package projection_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
	"github.com/alvadorncorp/gopackages/modeling/projection"
)

type salesSummary struct {
	created int
	paid    int
}

type orderEvent struct {
	created bool
	paid    bool
}

func projectSales(summary salesSummary, event orderEvent) salesSummary {
	if event.created {
		summary.created++
	}
	if event.paid {
		summary.paid++
	}
	return summary
}

var _ projection.Projection[salesSummary, orderEvent] = projectSales

func TestProjectionAdaptsAndInvokesOrdinaryFunction(t *testing.T) {
	t.Parallel()

	var project projection.Projection[salesSummary, orderEvent] = projectSales
	initial := salesSummary{created: 2, paid: 1}

	got := project(initial, orderEvent{paid: true})

	if got != (salesSummary{created: 2, paid: 2}) {
		t.Fatalf("projected summary = %+v, want created=2 paid=2", got)
	}
	if initial != (salesSummary{created: 2, paid: 1}) {
		t.Fatalf("initial summary = %+v, want unchanged", initial)
	}
}

func TestProjectionMayPreserveIrrelevantPrimitiveRepresentation(t *testing.T) {
	t.Parallel()

	projectPaid := projection.Projection[int, orderEvent](
		func(total int, event orderEvent) int {
			if event.paid {
				return total + 1
			}
			return total
		},
	)

	if got := projectPaid(7, orderEvent{created: true}); got != 7 {
		t.Fatalf("irrelevant event result = %d, want 7", got)
	}
}

func TestProjectionSupportsPersistentStateAndGenericEvent(t *testing.T) {
	t.Parallel()

	projectPrice := func(
		index immutable.Map[int, int],
		event map[string]int,
	) immutable.Map[int, int] {
		return index.Set(event["id"], event["price"])
	}
	var project projection.Projection[
		immutable.Map[int, int],
		map[string]int,
	] = projectPrice
	initial := immutable.Map[int, int]{}

	got := project(initial, map[string]int{"id": 42, "price": 199})

	if initial.Len() != 0 {
		t.Fatalf("initial index length = %d, want unchanged 0", initial.Len())
	}
	price, found := got.Get(42)
	if !found || price != 199 {
		t.Fatalf("projected price = %d, found=%t; want 199, true", price, found)
	}
}
