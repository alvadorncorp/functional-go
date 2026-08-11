package action_test

import (
	"context"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

const (
	chainDepth10  = 10
	chainDepth100 = 100
)

var (
	actionBenchSink      action.Action[int]
	actionValueBenchSink int
	actionErrorBenchSink error
)

func BenchmarkActionPure(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		actionBenchSink = action.Pure(answer)
	}
}

func BenchmarkActionRun(b *testing.B) {
	subject := action.Pure(answer)

	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionMap(b *testing.B) {
	subject := action.Map(action.Pure(answer), func(value int) int { return value + 1 })

	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionFlatMap(b *testing.B) {
	subject := action.FlatMap(action.Pure(answer), func(value int) action.Action[int] {
		return action.Pure(value + 1)
	})

	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionChain10(b *testing.B) {
	benchmarkActionChain(b, chainDepth10)
}

func BenchmarkActionChain100(b *testing.B) {
	benchmarkActionChain(b, chainDepth100)
}

func benchmarkActionChain(b *testing.B, depth int) {
	b.Helper()

	subject := action.Pure(0)
	for range depth {
		subject = action.Map(subject, func(value int) int { return value + 1 })
	}

	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}
