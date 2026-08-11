package transaction_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvadorncorp/gopackages/workflow/action"
	"github.com/alvadorncorp/gopackages/workflow/transaction"
)

// ExampleWithin shows the shape an adapter provides and the shape a workflow
// consumes: the adapter owns begin, commit, and rollback, and hands the
// workflow only a scope of capabilities.
func ExampleWithin() {
	insufficient := errors.New("insufficient balance")
	var log []string

	// The adapter. A real one would begin a database transaction here and build
	// the scope from its handle.
	boundary := func(ctx context.Context, fn func(context.Context, accounts) error) error {
		log = append(log, "BEGIN")
		scoped := accounts{
			Save: func(amount int) action.Action[struct{}] {
				return action.Do(func(context.Context) (struct{}, error) {
					log = append(log, fmt.Sprintf("UPDATE balance=%d", amount))
					return struct{}{}, nil
				})
			},
		}
		if err := fn(ctx, scoped); err != nil {
			log = append(log, "ROLLBACK")
			return err
		}
		log = append(log, "COMMIT")
		return nil
	}

	// The workflow. It never sees a driver type and never begins anything.
	commit := func(balance int) action.Action[int] {
		return transaction.Within(boundary, func(scoped accounts) action.Action[int] {
			if balance < 0 {
				return action.Fail[int](insufficient)
			}
			return action.Map(scoped.Save(balance), func(struct{}) int { return balance })
		})
	}

	balance, err := action.Run(context.Background(), commit(100))
	fmt.Println(balance, err, log)

	log = nil
	_, err = action.Run(context.Background(), commit(-1))
	fmt.Println(err, log)
	// Output:
	// 100 <nil> [BEGIN UPDATE balance=100 COMMIT]
	// insufficient balance [BEGIN ROLLBACK]
}
