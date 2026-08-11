package policy_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/primitives/predicate"
	"github.com/alvadorncorp/gopackages/modeling/policy"
)

type discountContext struct {
	VIP      bool
	Campaign bool
}

type discount struct{ Percentage int }

func isVIP(context discountContext) bool       { return context.VIP }
func hasCampaign(context discountContext) bool { return context.Campaign }

func ExamplePolicy_conditionalSelection() {
	vip := policy.When(
		predicate.Predicate[discountContext](isVIP),
		policy.Constant[discountContext](discount{Percentage: 15}),
	)
	campaign := policy.When(
		predicate.Predicate[discountContext](hasCampaign),
		policy.Constant[discountContext](discount{Percentage: 10}),
	)
	discountPolicy := policy.Fallback(
		policy.OrElse(vip, campaign),
		policy.Constant[discountContext](discount{}),
	)

	fmt.Println(discountPolicy(discountContext{VIP: true, Campaign: true}).Percentage)
	fmt.Println(discountPolicy(discountContext{Campaign: true}).Percentage)
	fmt.Println(discountPolicy(discountContext{}).Percentage)
	// Output:
	// 15
	// 10
	// 0
}

func ExampleCompose() {
	tier := policy.Policy[int, string](func(points int) string {
		if points >= 100 {
			return "gold"
		}
		return "standard"
	})
	creditLimit := policy.Policy[string, int](func(tier string) int {
		if tier == "gold" {
			return 5_000
		}
		return 1_000
	})

	fmt.Println(policy.Compose(tier, creditLimit)(120))
	// Output: 5000
}

func ExampleZip() {
	type pricingDecision struct {
		Discount int
		Shipping string
	}
	discountPolicy := policy.Policy[int, int](func(total int) int {
		if total >= 1_000 {
			return 10
		}
		return 0
	})
	shippingPolicy := policy.Policy[int, string](func(total int) string {
		if total >= 500 {
			return "express"
		}
		return "standard"
	})
	pricing := policy.Zip(discountPolicy, shippingPolicy, func(discount int, shipping string) pricingDecision {
		return pricingDecision{Discount: discount, Shipping: shipping}
	})

	fmt.Println(pricing(1_200))
	// Output: {10 express}
}

func ExampleChoose() {
	shipping := policy.Choose(
		predicate.Predicate[bool](func(international bool) bool { return international }),
		policy.Constant[bool]("international"),
		policy.Constant[bool]("domestic"),
	)

	fmt.Println(shipping(true))
	fmt.Println(shipping(false))
	// Output:
	// international
	// domestic
}
