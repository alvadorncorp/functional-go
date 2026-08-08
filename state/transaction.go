package state

// Transaction accumulates mutations against a captured State without
// materializing them.
//
// Begin freezes a snapshot of the parent State's base, plan head, and
// Version. Each Apply prepends one plan node on that working copy the same
// way State.Apply does. Commit returns the working State and freezes the
// transaction; it never calls Materialize.
//
// An empty transaction (Commit with no Apply) returns a State equivalent to
// the Begin snapshot: same base, plan, and Version.
//
// Transaction is single-use and not safe for concurrent use. After Commit,
// Apply and Commit panic with "state: transaction used after Commit".
// Begin, Apply, and Commit never mutate the parent State value that Begin
// was called on.
type Transaction[T any] struct {
	working   State[T]
	committed bool
}

// Begin captures s as the base of a new Transaction. The receiver is
// unchanged.
//
// Mutations recorded on the transaction build on s's plan. Commit returns a
// new State with those mutations applied to the plan; Version advances by
// one per Apply, matching sequential State.Apply. Commit is lazy and does
// not call Materialize.
func (s State[T]) Begin() *Transaction[T] {
	return &Transaction[T]{working: s}
}

// Apply records m on the transaction's working plan without invoking m.
// The parent State from Begin is unchanged.
//
// Apply returns the receiver so callers may chain Apply calls. After
// Commit, Apply panics with "state: transaction used after Commit".
// Transaction is not concurrency-safe.
func (tx *Transaction[T]) Apply(m Mutation[T]) *Transaction[T] {
	if tx.committed {
		panic(panicTransactionUsedAfterCommit)
	}
	tx.working = tx.working.Apply(m)
	return tx
}

// Commit returns a State that includes every mutation applied on tx since
// Begin, layered on the captured plan. Commit does not Materialize and does
// not mutate the parent State from Begin.
//
// If no Apply ran, Commit returns a State equivalent to the Begin snapshot
// (same base, plan, and Version). After Commit, further Apply or Commit
// panics with "state: transaction used after Commit". Transaction is not
// concurrency-safe.
func (tx *Transaction[T]) Commit() State[T] {
	if tx.committed {
		panic(panicTransactionUsedAfterCommit)
	}
	tx.committed = true
	return tx.working
}
