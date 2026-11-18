package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func testConsumer(store *Store) Consumer {
	return Consumer{
		CreateOrder: CreateOrderWorkflow(store.Dependencies()),
		Backoff:     func(int) time.Duration { return 0 },
	}
}

func testMessage(orderID string) OrderMessage {
	return OrderMessage{OrderID: orderID, Customer: "cust-1", Product: "prod-1", Units: 2}
}

// A transient failure inside the transaction must roll the whole attempt back
// and leave the store clean, so the retry can commit exactly one order.
func TestTransientFailureRollsBackAndRetriesToExactlyOneOrder(t *testing.T) {
	t.Parallel()

	store := NewStore()
	store.FailNextOutboxAppends(1)
	consumer := testConsumer(store)

	if err := consumer.Handle(t.Context(), testMessage("ord-1")); err != nil {
		t.Fatalf("Handle() error = %v, want nil", err)
	}

	trace := store.Trace()
	if got := strings.Count(trace, "transaction"); got != 2 {
		t.Fatalf("transaction count = %d, want 2 (one rolled back, one committed)", got)
	}
	if got := strings.Count(trace, "rollback"); got != 1 {
		t.Fatalf("rollback count = %d, want 1", got)
	}
	orders, events := store.Totals()
	if orders != 1 || events != 1 {
		t.Fatalf("totals = (%d orders, %d events), want (1, 1)", orders, events)
	}
}

// The broker may redeliver the same message at any time. The retry around the
// workflow plus the store's upsert must keep one order and one event.
func TestRedeliveryUpsertsInsteadOfDuplicating(t *testing.T) {
	t.Parallel()

	store := NewStore()
	consumer := testConsumer(store)
	message := testMessage("ord-2")

	if err := consumer.Handle(t.Context(), message); err != nil {
		t.Fatalf("first Handle() error = %v, want nil", err)
	}
	store.ResetTrace()
	if err := consumer.Handle(t.Context(), message); err != nil {
		t.Fatalf("redelivered Handle() error = %v, want nil", err)
	}

	trace := store.Trace()
	if !strings.Contains(trace, "save-order") {
		t.Fatalf("trace = %q, want a save-order step", trace)
	}
	if strings.Contains(trace, "append-outbox") {
		t.Fatalf("trace = %q, want no second outbox append on redelivery", trace)
	}
	orders, events := store.Totals()
	if orders != 1 || events != 1 {
		t.Fatalf("totals = (%d orders, %d events), want (1, 1)", orders, events)
	}
}

// Domain rejections are not transient: the consumer must not retry them, and
// the rejected decision must never open a transaction.
func TestDomainRejectionIsNotRetried(t *testing.T) {
	t.Parallel()

	store := NewStore()
	consumer := testConsumer(store)
	message := OrderMessage{OrderID: "ord-3", Customer: "cust-1", Product: "prod-1", Units: 0}

	err := consumer.Handle(t.Context(), message)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("Handle() error = %v, want ErrInvalidQuantity identity", err)
	}

	trace := store.Trace()
	if got := strings.Count(trace, "load-customer"); got != 1 {
		t.Fatalf("load-customer count = %d, want 1 (no retry for domain rejections)", got)
	}
	if strings.Contains(trace, "transaction") {
		t.Fatalf("trace = %q, want no transaction for a rejected decision", trace)
	}
}
