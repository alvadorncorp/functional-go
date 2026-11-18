package main

import (
	"context"
	"errors"
	"slices"
	"testing"

	"go.alvadorn.org/functional/values/optional"
	"go.alvadorn.org/functional/workflow/action"
)

// TestOutboxFailureRollsBackAndRetries proves the rollback leaves no partial
// writes behind and that a retry succeeds once the one-shot failure is
// consumed.
func TestOutboxFailureRollsBackAndRetries(t *testing.T) {
	rt := newDemoRuntime()
	rt.store.FailNext(failOutbox)
	request := CheckoutRequest{
		OrderID:    demoOrderID,
		CustomerID: demoCustomerID,
		CouponCode: optional.Some(demoCouponCode),
		Payment:    demoImmediate(),
	}

	if _, err := action.Run(context.Background(), rt.checkout(request)); !errors.Is(err, errInjectedFailure) {
		t.Fatalf("error = %v, want the injected outbox failure", err)
	}
	snap := rt.store.Snapshot()
	if got := snap.Orders[demoOrderID].Status(); got != OrderStatusDraft {
		t.Fatalf("status = %s, want draft after rollback", got)
	}
	if got := snap.Coupons[demoCouponCode].RemainingUses; got != demoCouponUses {
		t.Fatalf("coupon uses = %d, want %d after rollback", got, demoCouponUses)
	}
	if len(snap.Payments) != 0 || len(snap.Outbox) != 0 {
		t.Fatalf("payments = %d, outbox = %d, want both empty after rollback", len(snap.Payments), len(snap.Outbox))
	}

	committed, err := action.Run(context.Background(), rt.checkout(request))
	if err != nil {
		t.Fatalf("retry error = %v", err)
	}
	if got := committed.Order.Status(); got != OrderStatusCheckedOut {
		t.Fatalf("status after retry = %s, want checked_out", got)
	}
	if got := committed.Order.Total().Cents(); got != paidCents {
		t.Fatalf("total after retry = %d, want %d", got, paidCents)
	}
	snap = rt.store.Snapshot()
	if len(snap.Outbox) != retryOutboxCount {
		t.Fatalf("outbox = %d, want %d after retry", len(snap.Outbox), retryOutboxCount)
	}
}

const (
	paidCents        = int64(9_000)
	retryOutboxCount = 2
)

// TestPremiumLoyaltyReplaysPromotion proves the loyalty branch: the discount
// event carries the promotion id, replay records it on the rebuilt order, and
// the projection counts the reason.
func TestPremiumLoyaltyReplaysPromotion(t *testing.T) {
	seed := demoDraft(demoOrderID)
	premium := demoCustomer()
	premium.Tier = CustomerPremium
	premium.CompletedOrders = minimumCompletedOrders

	decision, err := Decide(seed, &DecideInput{
		Customer: premium,
		Payment:  demoImmediate(),
		Coupon:   optional.None[Coupon](),
	})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	events := slices.Collect(decision.Events())

	rebuilt := ReplayOrder(seed, slices.Values(events))
	if rebuilt.PromotionsCount() != 1 {
		t.Fatalf("promotions = %d, want the premium-loyalty promotion", rebuilt.PromotionsCount())
	}
	summary := ProjectSummary(slices.Values(events))
	if summary.DiscountCount != 1 {
		t.Fatalf("discounts = %d, want 1", summary.DiscountCount)
	}
	amount, ok := summary.Reasons().Get(reasonPremiumLoyalty)
	if !ok || amount.Cents() != loyaltyDiscountCents {
		t.Fatalf("loyalty reason = (%v, %v), want %d cents", amount, ok, loyaltyDiscountCents)
	}
}

// loyaltyDiscountCents is five percent of the seeded item.
const loyaltyDiscountCents = int64(500)
