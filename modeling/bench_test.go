package modeling_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

var (
	decisionBenchSink modeling.Decision[int, int]
	decisionEventSink int
)

func BenchmarkDecisionEmpty(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		decisionBenchSink = modeling.Decide[int](0)
	}
}

func BenchmarkDecisionEmitOne(b *testing.B) {
	base := modeling.Decide[int](0)
	b.ReportAllocs()
	for b.Loop() {
		decisionBenchSink = base.Emit(1)
	}
}

func BenchmarkDecisionEmitBatch(b *testing.B) {
	base := modeling.Decide[int](0)
	b.ReportAllocs()
	for b.Loop() {
		decisionBenchSink = base.Emit(1, 2, 3, 4, 5, 6, 7, 8)
	}
}

func BenchmarkDecisionEmitChain(b *testing.B) {
	base := modeling.Decide[int](0)
	b.ReportAllocs()
	for b.Loop() {
		decision := base
		for event := range 8 {
			decision = decision.Emit(event)
		}
		decisionBenchSink = decision
	}
}

func BenchmarkDecisionEvents(b *testing.B) {
	decision := modeling.Decide[int](0).Emit(1, 2, 3, 4, 5, 6, 7, 8)
	b.ReportAllocs()
	for b.Loop() {
		total := 0
		for event := range decision.Events() {
			total += event
		}
		decisionEventSink = total
	}
}
