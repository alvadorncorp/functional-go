package archtest

import (
	"context"
	"time"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/plan"
	"go.alvadorn.org/functional/workflow/resilience"
)

var (
	_ plan.Kind                    = plan.KindEmpty
	_ func(plan.Kind) string       = plan.Kind.String
	_ func(string) plan.Plan       = plan.Step
	_ func(...plan.Plan) plan.Plan = plan.Sequence
	_ func(...plan.Plan) plan.Plan = plan.Transaction
	_ func(plan.Plan) plan.Kind    = plan.Plan.Kind
	_ func(plan.Plan) string       = plan.Plan.Name
	_ func(plan.Plan) []plan.Plan  = plan.Plan.Children
	_ func(plan.Plan) string       = plan.Plan.String
	_ func(
		func(context.Context) error,
	) action.Action[struct{}] = action.Exec
	_ func(
		workflow.Workflow[int, string],
		int,
	) action.Action[string] = workflow.Provide[int, string]
	_ func(
		workflow.Workflow[string, int],
		func(bool) string,
	) workflow.Workflow[bool, int] = workflow.Contramap[bool, string, int]
	_ func(
		action.Action[int],
		action.Action[string],
	) action.Action[action.Pair[int, string]] = action.Zip[int, string]
	_ func(
		action.Action[int],
		action.Action[string],
		func(int, string) bool,
	) action.Action[bool] = action.Both[int, string, bool]
	_ func(
		action.Action[int],
		func(int) action.Action[struct{}],
	) action.Action[int] = action.Tap[int]
	_ func(
		[]int,
		func(int) action.Action[string],
	) action.Action[[]string] = action.Traverse[int, string]
	_ func(action.Action[int]) action.Action[action.AttemptResult[int]] = action.Attempt[int]
	_ func(
		action.Action[int],
		action.Action[struct{}],
	) action.Action[int] = action.Ensure[int]
	_ func(
		action.Action[string],
		func(string) action.Action[int],
		func(string) action.Action[struct{}],
	) action.Action[int] = action.Bracket[string, int]
	_ func(
		action.Action[string],
		func(string) action.Action[int],
		func(string) action.Action[struct{}],
	) action.Action[int] = action.AcquireUseRelease[string, int]
	_ func(resilience.RetryPolicy, action.Action[int]) action.Action[int] = resilience.Retry[int]
	_ func(time.Duration, action.Action[int]) action.Action[int]          = resilience.Timeout[int]
)
