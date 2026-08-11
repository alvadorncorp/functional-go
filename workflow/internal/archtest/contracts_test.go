package archtest

import (
	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

var (
	_ func(workflow.Workflow[int, string], int) action.Action[string]                             = workflow.Provide[int, string]
	_ func(workflow.Workflow[string, int], func(bool) string) workflow.Workflow[bool, int]        = workflow.Contramap[bool, string, int]
	_ func(action.Action[int], action.Action[string]) action.Action[action.Pair[int, string]]     = action.Zip[int, string]
	_ func(action.Action[int], action.Action[string], func(int, string) bool) action.Action[bool] = action.Both[int, string, bool]
	_ func(action.Action[int], func(int) action.Action[struct{}]) action.Action[int]              = action.Tap[int]
	_ func(action.Action[int]) action.Action[action.AttemptResult[int]]                           = action.Attempt[int]
)
