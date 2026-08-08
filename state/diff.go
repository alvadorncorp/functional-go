package state

// Diff is a structural plan of mutations between two States of the same
// lineage. It stores the steps that take From's plan to To's plan without
// materializing either side.
//
// The zero Diff is empty: From and To report 0, Len is 0, and Apply accepts
// only a zero-lineage base at Version 0 with a nil plan head (the zero
// State).
type Diff[T any] struct {
	line     *lineage
	fromVer  Version
	toVer    Version
	fromHead *planNode[T]
	steps    []Mutation[T]
}

// DiffBetween returns the plan steps that evolve from into to.
//
// from and to must share a lineage token (pointer equality). Both zero
// States (nil lineage) yield an empty Diff. A single nil lineage against a
// non-nil one panics with "state: unrelated lineages".
//
// to must be a plan descendant of from: walking to's reverse-linked plan
// must eventually reach from's head, or from must be an empty-plan root of
// the shared lineage (head nil). Sibling branches and reverse ancestry
// panic with "state: unrelated lineages".
//
// The returned Diff does not call Materialize. Len is the number of
// mutations strictly after from's head up to and including to's head.
func DiffBetween[T any](from, to State[T]) Diff[T] {
	if from.line == nil && to.line == nil {
		// Independent zero States share the nil token; empty Diff only.
		return Diff[T]{}
	}
	requireSameLineage(from.line, to.line)

	if from.head == to.head && from.ver == to.ver {
		return Diff[T]{
			line:     from.line,
			fromVer:  from.ver,
			toVer:    to.ver,
			fromHead: from.head,
		}
	}

	requirePlanDescendant(from, to)

	return Diff[T]{
		line:     from.line,
		fromVer:  from.ver,
		toVer:    to.ver,
		fromHead: from.head,
		steps:    collectStepsAfter(from.head, to.head),
	}
}

func requireSameLineage(a, b *lineage) {
	if a != b {
		panic(panicUnrelatedLineages)
	}
}

// requirePlanDescendant panics unless to's plan is a descendant of from's
// plan within a shared lineage (caller already checked lineage equality).
func requirePlanDescendant[T any](from, to State[T]) {
	if from.head == nil {
		// Empty-plan root of this lineage: every step on to is in the diff.
		if from.ver > to.ver {
			panic(panicUnrelatedLineages)
		}
		return
	}
	for p := to.head; p != nil; p = p.older {
		if p == from.head {
			return
		}
	}
	panic(panicUnrelatedLineages)
}

// collectStepsAfter returns mutations strictly after fromHead up to and
// including toHead, in oldest→newest order.
func collectStepsAfter[T any](fromHead, toHead *planNode[T]) []Mutation[T] {
	n := 0
	for p := toHead; p != nil && p != fromHead; p = p.older {
		n++
	}
	steps := make([]Mutation[T], n)
	i := n - 1
	for p := toHead; p != nil && p != fromHead; p = p.older {
		steps[i] = p.mut
		i--
	}
	return steps
}

// From reports the Version of the from State passed to DiffBetween.
func (d Diff[T]) From() Version {
	return d.fromVer
}

// To reports the Version of the to State passed to DiffBetween.
func (d Diff[T]) To() Version {
	return d.toVer
}

// Len reports how many mutations the Diff records (oldest→newest).
func (d Diff[T]) Len() int {
	return len(d.steps)
}

// Apply replays the Diff's mutations onto base, returning the resulting
// State without calling Materialize.
//
// base must share the Diff's lineage, report Version equal to From, and
// hold the same plan head DiffBetween captured for from. Any mismatch
// panics with "state: unrelated lineages".
func (d Diff[T]) Apply(base State[T]) State[T] {
	if base.line != d.line || base.ver != d.fromVer || base.head != d.fromHead {
		panic(panicUnrelatedLineages)
	}
	s := base
	for _, m := range d.steps {
		s = s.Apply(m)
	}
	return s
}
