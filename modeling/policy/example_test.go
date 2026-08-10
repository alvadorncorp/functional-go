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

type discount struct {
	Percentage int
}

func isVIP(context discountContext) bool { return context.VIP }

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
		policy.FirstApplicable(vip, campaign),
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
