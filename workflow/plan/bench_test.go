package plan_test

import (
	"testing"

	"go.alvadorn.org/functional/workflow/plan"
)

var (
	planBenchSink         plan.Plan
	planChildrenBenchSink []plan.Plan
)

func BenchmarkStepConstruction(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		planBenchSink = plan.Step("load-customer")
	}
}

func BenchmarkSequenceConstruction(b *testing.B) {
	a, c := plan.Step("a"), plan.Step("c")
	b.ReportAllocs()
	for b.Loop() {
		planBenchSink = plan.Sequence(a, c)
	}
}

func BenchmarkChildrenInspection(b *testing.B) {
	subject := plan.Sequence(plan.Step("a"), plan.Step("b"))
	b.ReportAllocs()
	for b.Loop() {
		planChildrenBenchSink = subject.Children()
	}
}
