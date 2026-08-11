package transaction_test

import (
	"context"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
	"github.com/alvadorncorp/gopackages/workflow/transaction"
)

var (
	transactionValueSink int
	transactionErrorSink error
)

func benchBoundary() transaction.Boundary[accounts] {
	writable := accounts{
		Save: func(int) action.Action[struct{}] { return action.Pure(struct{}{}) },
	}
	return func(ctx context.Context, fn func(context.Context, accounts) error) error {
		return fn(ctx, writable)
	}
}

func BenchmarkWithin(b *testing.B) {
	program := transaction.Within(benchBoundary(), func(scoped accounts) action.Action[int] {
		return action.Map(scoped.Save(savedAmount), func(struct{}) int { return savedAmount })
	})

	b.ReportAllocs()
	for b.Loop() {
		transactionValueSink, transactionErrorSink = action.Run(context.Background(), program)
	}
}

func BenchmarkWithinNestedActions(b *testing.B) {
	program := transaction.Within(benchBoundary(), func(scoped accounts) action.Action[int] {
		return action.FlatMap(scoped.Save(savedAmount), func(struct{}) action.Action[int] {
			return action.Map(scoped.Save(savedAmount), func(struct{}) int { return savedAmount })
		})
	})

	b.ReportAllocs()
	for b.Loop() {
		transactionValueSink, transactionErrorSink = action.Run(context.Background(), program)
	}
}
