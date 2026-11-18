// Command checkout is a memory-backed demo of the functional checkout
// workflow: one successful purchase, proof that an outbox failure rolls the
// whole transaction back, and a replay/projection pass over the emitted
// events. It needs no network, database, or environment.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"go.alvadorn.org/functional/values/either"
	"go.alvadorn.org/functional/values/optional"
	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
)

const (
	demoOrderID       = OrderID("ord-1")
	rollbackOrderID   = OrderID("ord-2")
	demoCustomerID    = CustomerID("cust-1")
	demoProductID     = ProductID("prod-1")
	demoCouponCode    = CouponCode("SAVE10")
	demoItemCents     = int64(10_000)
	demoCouponPercent = uint8(10)
	demoCouponUses    = uint32(5)
	demoPaymentMethod = PaymentMethodID("card")
	demoDeferredDays  = uint16(30)
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "checkout demo failed: %v\n", err)
		os.Exit(1)
	}
}

// demoRuntime wires one seeded store and the shared checkout workflow.
type demoRuntime struct {
	store    *Store
	checkout workflow.Workflow[CheckoutRequest, CheckoutResult]
}

func newDemoRuntime() *demoRuntime {
	store := NewStore(
		[]Order{demoDraft(demoOrderID), demoDraft(rollbackOrderID)},
		[]CustomerProfile{demoCustomer()},
		[]Coupon{demoCoupon()},
	)
	return &demoRuntime{
		store: store,
		checkout: Checkout(CheckoutDependencies{
			Orders:    store.Orders(),
			Customers: store.Customers(),
			Coupons:   store.Coupons(),
			Boundary:  store.Boundary(),
		}),
	}
}

// demoDraft builds the single-item draft every scenario starts from.
func demoDraft(id OrderID) Order {
	order := NewOrder(id, demoCustomerID, USD)
	item := OrderItem{
		ProductID: demoProductID,
		UnitPrice: NewMoney(demoItemCents, USD).MustGet(),
		Quantity:  NewQuantity(1).MustGet(),
	}
	return order.AddItem(item).MustGet()
}

// demoCustomer is standard tier with too few completed orders for the
// loyalty discount.
func demoCustomer() CustomerProfile {
	return CustomerProfile{ID: demoCustomerID, Tier: CustomerStandard, CompletedOrders: 2}
}

func demoCoupon() Coupon {
	return NewCoupon(demoCouponCode, demoCouponPercent, false, demoCouponUses).MustGet()
}

func demoImmediate() PaymentChoice {
	return ImmediatePaymentChoice(ImmediatePayment{Method: demoPaymentMethod})
}

// run executes the three-part demo: one successful checkout, an injected
// outbox failure that rolls a second checkout back, and a replay/projection
// pass proving the committed state is the pure decision state.
func run(ctx context.Context, writer io.Writer) error {
	rt := newDemoRuntime()

	result, err := action.Run(ctx, rt.checkout(CheckoutRequest{
		OrderID:    demoOrderID,
		CustomerID: demoCustomerID,
		CouponCode: optional.Some(demoCouponCode),
		Payment:    demoImmediate(),
	}))
	if err != nil {
		return err
	}
	payments := rt.store.Snapshot().Payments
	if printErr := printCheckout(writer, payments, result); printErr != nil {
		return printErr
	}
	if rollErr := proveRollback(ctx, writer, rt); rollErr != nil {
		return rollErr
	}
	return proveReplay(writer, result)
}

// printCheckout reports the committed purchase and its outbox rows in
// emission order.
func printCheckout(writer io.Writer, payments map[OrderID]PaymentRecord, result CheckoutResult) error {
	order := result.Order
	if _, err := fmt.Fprintf(writer,
		"checkout ok order_id=%s total_cents=%d currency=%s payment=%s outbox=%d\n",
		order.ID(), order.Total().Cents(), order.Total().Currency(),
		paymentMode(payments, order.ID()), result.Outbox.Len(),
	); err != nil {
		return err
	}
	for i, message := range result.Outbox.Items() {
		if _, err := fmt.Fprintf(writer, "  outbox[%d] %s aggregate=%s\n", i, message.Type, message.AggregateID); err != nil {
			return err
		}
	}
	return nil
}

// paymentMode names the committed payment branch for one order.
func paymentMode(payments map[OrderID]PaymentRecord, id OrderID) string {
	record, ok := payments[id]
	if !ok {
		return "unknown"
	}
	return either.Match(record.Payment,
		func(PaymentMethodID) string { return "immediate" },
		func(uint16) string { return "deferred" },
	)
}

// proveRollback arms an outbox failure and attempts a second checkout with a
// deferred payment and the same coupon. The injected error must discard every
// staged write: the order stays draft, the coupon keeps its uses, and no
// payment or outbox row commits.
func proveRollback(ctx context.Context, writer io.Writer, rt *demoRuntime) error {
	rt.store.FailNext(failOutbox)
	_, err := action.Run(ctx, rt.checkout(CheckoutRequest{
		OrderID:    rollbackOrderID,
		CustomerID: demoCustomerID,
		CouponCode: optional.Some(demoCouponCode),
		Payment:    DeferredPaymentChoice(DeferredPayment{Days: demoDeferredDays}),
	}))
	if err == nil {
		return errors.New("expected the injected outbox failure")
	}
	if !errors.Is(err, errInjectedFailure) {
		return fmt.Errorf("unexpected checkout failure: %w", err)
	}
	snap := rt.store.Snapshot()
	order := snap.Orders[rollbackOrderID]
	_, writeErr := fmt.Fprintf(writer,
		"rollback ok err=%q order=%s status=%s coupon_uses=%d payments=%d outbox=%d\n",
		err, rollbackOrderID, order.Status(),
		snap.Coupons[demoCouponCode].RemainingUses, len(snap.Payments), len(snap.Outbox),
	)
	return writeErr
}

// proveReplay re-decides the seeded draft purely, replays the events back
// onto it, and projects them into the read model. Replaying must rebuild
// exactly the state the workflow committed, and the projection must see the
// same events.
func proveReplay(writer io.Writer, committed CheckoutResult) error {
	seed := demoDraft(demoOrderID)
	decision, err := Decide(seed, &DecideInput{
		Customer: demoCustomer(),
		Payment:  demoImmediate(),
		Coupon:   optional.Some(demoCoupon()),
	})
	if err != nil {
		return err
	}
	events := slices.Collect(decision.Events())
	rebuilt := ReplayOrder(seed, slices.Values(events))
	if !rebuilt.SameState(committed.Order) {
		return errors.New("replay rebuilt different state than the workflow committed")
	}
	summary := ProjectSummary(slices.Values(events))
	if _, printErr := fmt.Fprintf(writer,
		"replay ok events=%d version=%d total_cents=%d discount_cents=%d promotions=%d\n",
		len(events), rebuilt.Version(), rebuilt.Total().Cents(),
		rebuilt.Discount().Cents(), rebuilt.PromotionsCount(),
	); printErr != nil {
		return printErr
	}
	_, err = fmt.Fprintf(writer,
		"projection ok events=%d discounts=%d reasons=%s\n",
		summary.EventCount, summary.DiscountCount, reasonList(summary),
	)
	return err
}

// reasonList renders accumulated discount reasons sorted by name, or "none".
func reasonList(summary Summary) string {
	reasons := summary.Reasons().Items()
	keys := slices.Sorted(maps.Keys(reasons))
	if len(keys) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, reasons[key].Cents()))
	}
	return strings.Join(parts, ",")
}
