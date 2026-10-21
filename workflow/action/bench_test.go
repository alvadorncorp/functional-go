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
	actionBenchSink        action.Action[int]
	actionValueBenchSink   int
	actionErrorBenchSink   error
	actionPairBenchSink    action.Pair[int, int]
	actionAttemptBenchSink action.AttemptResult[int]
	actionSliceBenchSink   []int
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

func BenchmarkActionZip(b *testing.B) {
	subject := action.Zip(action.Pure(20), action.Pure(22))
	b.ReportAllocs()
	for b.Loop() {
		actionPairBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionZipManual(b *testing.B) {
	subject := action.FlatMap(action.Pure(20), func(first int) action.Action[action.Pair[int, int]] {
		return action.Map(action.Pure(22), func(second int) action.Pair[int, int] {
			return action.Pair[int, int]{First: first, Second: second}
		})
	})
	b.ReportAllocs()
	for b.Loop() {
		actionPairBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionBoth(b *testing.B) {
	subject := action.Both(action.Pure(20), action.Pure(22), func(a, c int) int { return a + c })
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionBothManual(b *testing.B) {
	subject := action.FlatMap(action.Pure(20), func(first int) action.Action[int] {
		return action.Map(action.Pure(22), func(second int) int { return first + second })
	})
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionTap(b *testing.B) {
	subject := action.Tap(action.Pure(answer), func(int) action.Action[struct{}] {
		return action.Pure(struct{}{})
	})
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionTapManual(b *testing.B) {
	effect := func(int) action.Action[struct{}] { return action.Pure(struct{}{}) }
	subject := action.FlatMap(action.Pure(answer), func(value int) action.Action[int] {
		return action.Map(effect(value), func(struct{}) int { return value })
	})
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionTraverse(b *testing.B) {
	items := benchmarkTraverseItems()
	subject := action.Traverse(items, func(item int) action.Action[int] { return action.Pure(item + 1) })
	b.ReportAllocs()
	for b.Loop() {
		actionSliceBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionTraverseManual(b *testing.B) {
	items := benchmarkTraverseItems()
	binder := func(item int) action.Action[int] { return action.Pure(item + 1) }
	subject := action.Do(func(ctx context.Context) ([]int, error) {
		values := make([]int, 0, len(items))
		for _, item := range items {
			value, err := binder(item)(ctx)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	})
	b.ReportAllocs()
	for b.Loop() {
		actionSliceBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func benchmarkTraverseItems() []int {
	items := make([]int, chainDepth10)
	for i := range items {
		items[i] = i
	}
	return items
}

func BenchmarkActionAttempt(b *testing.B) {
	subject := action.Attempt(action.Pure(answer))
	b.ReportAllocs()
	for b.Loop() {
		actionAttemptBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionEnsure(b *testing.B) {
	subject := action.Ensure(action.Pure(answer), action.Pure(struct{}{}))
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionBracket(b *testing.B) {
	subject := action.Bracket(
		action.Pure(answer),
		action.Pure[int],
		func(int) action.Action[struct{}] { return action.Pure(struct{}{}) },
	)
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionAcquireUseRelease(b *testing.B) {
	subject := action.AcquireUseRelease(
		action.Pure(answer),
		action.Pure[int],
		func(int) action.Action[struct{}] { return action.Pure(struct{}{}) },
	)
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}

func BenchmarkActionBracketManual(b *testing.B) {
	acquire := action.Pure(answer)
	use := action.Pure[int]
	release := func(int) action.Action[struct{}] { return action.Pure(struct{}{}) }
	subject := action.Do(func(ctx context.Context) (int, error) {
		resource, err := acquire(ctx)
		if err != nil {
			return 0, err
		}
		value, useErr := use(resource)(ctx)
		_, releaseErr := release(resource)(ctx)
		if useErr != nil {
			return 0, useErr
		}
		if releaseErr != nil {
			return 0, releaseErr
		}
		return value, nil
	})
	b.ReportAllocs()
	for b.Loop() {
		actionValueBenchSink, actionErrorBenchSink = action.Run(context.Background(), subject)
	}
}
