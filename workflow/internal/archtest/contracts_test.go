package archtest

import (
	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

var (
	_ func(workflow.Workflow[int, string], int) action.Action[string]                      = workflow.Provide[int, string]
	_ func(workflow.Workflow[string, int], func(bool) string) workflow.Workflow[bool, int] = workflow.Contramap[bool, string, int]
)
