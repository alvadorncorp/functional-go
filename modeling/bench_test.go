package modeling_test

import (
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

var (
	decisionBenchSink       modeling.Decision[int, int]
	decisionStringBenchSink modeling.Decision[int, string]
	decisionEventSink       int
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

var decisionBenchmarkSizes = [...]int{0, 1, 8, 32, 128, 1024}

func decisionBenchmarkEvents(size int) []int {
	events := make([]int, size)
	for index := range events {
		events[index] = index
	}
	return events
}

func BenchmarkDecisionEmit(b *testing.B) {
	for _, size := range decisionBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := decisionBenchmarkEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				decision := modeling.Decide[int](0)
				for _, event := range events {
					decision = modeling.Emit(decision, event)
				}
				decisionBenchSink = decision
			}
		})
	}
}

func BenchmarkDecisionAppendEvents(b *testing.B) {
	for _, size := range decisionBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			base := modeling.Decide[int](0)
			events := decisionBenchmarkEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				decisionBenchSink = modeling.AppendEvents(base, events...)
			}
		})
	}
}

func BenchmarkDecisionCombine(b *testing.B) {
	for _, size := range decisionBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := decisionBenchmarkEvents(size)
			middle := len(events) / 2
			left := modeling.AppendEvents(modeling.Decide[int](1), events[:middle]...)
			right := modeling.AppendEvents(modeling.Decide[int](2), events[middle:]...)
			b.ReportAllocs()
			for b.Loop() {
				decisionBenchSink = modeling.Combine(left, right)
			}
		})
	}
}

func BenchmarkDecisionMapEvents(b *testing.B) {
	for _, size := range decisionBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := decisionBenchmarkEvents(size)
			decision := modeling.AppendEvents(modeling.Decide[int](0), events...)
			b.ReportAllocs()
			for b.Loop() {
				decisionStringBenchSink = modeling.MapEvents(decision, strconv.Itoa)
			}
		})
	}
}
