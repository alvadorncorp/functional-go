package state

// Mutation is a pure evolution step over a value of type T.
//
// Callers must treat m as free of side effects on shared state: Apply only
// records the function, and Materialize may invoke it any number of times.
type Mutation[T any] func(T) T

// Version is a monotonic lineage counter for a State value.
//
// Of and the zero State report Version 0. Each Apply returns a new State
// whose Version is the receiver's Version plus one. Sibling branches from
// the same parent share that parent's Version and both advance by one
// independently; Version does not encode plan shape.
type Version uint64

// lineage is an unexported identity token shared by every State produced
// from the same Of root. DiffBetween and Diff.Apply compare lineage by
// pointer equality so independent Of roots never join even when their base
// values compare equal. It must remain non-zero-sized: Go is permitted to
// coalesce pointers to distinct zero-sized allocations.
type lineage struct{ _ byte }

// zeroValueLineage gives all zero States of a given T their canonical empty
// root. A zero State has no storage for a per-value token, so the first Apply
// must use this stable token for diffs from that zero root to remain valid.
// States of different T cannot be mixed by the type system.
var zeroValueLineage = new(lineage)

// State is an immutable plan of pure evolutions over a base value of type T.
//
// The zero State is valid: Materialize returns the zero value of T with no
// mutations applied, and Version reports 0. Apply never mutates its
// receiver; it returns a new State that shares the prior reverse-linked
// plan and prepends one node. A State derived from the zero value uses the
// canonical zero-value lineage, so it can be diffed from that empty root.
//
// Mutations are not executed by Apply. Materialize walks the plan from
// oldest to newest, re-applying every step on each call (no cache).
//
// Immutability of the payload is only as deep as T itself: State[[]byte]
// shares the slice's backing array across versions unless T is itself an
// immutable type.
type State[T any] struct {
	base T
	head *planNode[T]
	ver  Version
	line *lineage
}

// Of returns a State holding value with an empty plan at Version 0.
//
// Each call allocates a fresh lineage token. States derived from that root
// via Apply (or Transaction) share the token; DiffBetween panics across
// independent Of roots.
func Of[T any](value T) State[T] {
	return State[T]{base: value, line: new(lineage)}
}

// Version reports the lineage counter for s.
//
// The zero State and values from Of report 0. Each Apply advances the
// counter by one on the returned State only.
func (s State[T]) Version() Version {
	return s.ver
}

// Apply returns a new State that records m after every mutation already
// planned on s. The receiver is unchanged. m is not invoked.
//
// Apply is O(1): one plan node is allocated and linked as the new head.
// The returned State's Version is s.Version() + 1. The lineage token is
// preserved, or the canonical zero-value token is selected for a zero
// receiver.
func (s State[T]) Apply(m Mutation[T]) State[T] {
	return State[T]{
		base: s.base,
		head: &planNode[T]{older: s.head, mut: m},
		ver:  s.ver + 1,
		line: s.lineageToken(),
	}
}

func (s State[T]) lineageToken() *lineage {
	if s.line == nil {
		return zeroValueLineage
	}
	return s.line
}

// Materialize returns the value obtained by applying every planned mutation
// to the base, in oldest-to-newest order.
//
// Each call re-applies the full plan; there is no cached result. A zero
// State yields the zero value of T.
func (s State[T]) Materialize() T {
	n := 0
	for p := s.head; p != nil; p = p.older {
		n++
	}
	if n == 0 {
		return s.base
	}

	// Collect oldest→newest so a single forward pass applies in plan order.
	steps := make([]Mutation[T], n)
	i := n - 1
	for p := s.head; p != nil; p = p.older {
		steps[i] = p.mut
		i--
	}

	v := s.base
	for _, m := range steps {
		v = m(v)
	}
	return v
}

// Snapshot is a frozen materialization of a State at a point in time.
//
// Snapshot materializes once: Value is the result of Materialize at the
// time Snapshot is called, and Version is the State's Version then.
// Subsequent Apply on the source State (or any other evolution) does not
// change an already-taken Snapshot.
type Snapshot[T any] struct {
	Version Version
	Value   T
}

// Snapshot materializes s once and freezes Version and Value.
//
// The returned Snapshot is independent of further Apply on s: evolving s
// yields a new State and leaves the Snapshot's fields unchanged.
func (s State[T]) Snapshot() Snapshot[T] {
	return Snapshot[T]{
		Version: s.ver,
		Value:   s.Materialize(),
	}
}
