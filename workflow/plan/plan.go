// Package plan provides immutable, non-executable descriptions of workflow
// structure. A Plan is application-declared data: it never receives a context,
// runs an action, opens a transaction, or performs I/O.
package plan

import "slices"

// Kind identifies the structural form of a Plan.
type Kind uint8

const (
	KindEmpty Kind = iota
	KindStep
	KindSequence
	KindTransaction
)

// Plan is an immutable structural description. Its zero value is KindEmpty.
type Plan struct {
	kind     Kind
	name     string
	children []Plan
}

// Step describes one named conceptual operation.
func Step(name string) Plan {
	if name == "" {
		panic(panicEmptyStepName)
	}
	return Plan{kind: KindStep, name: name}
}

// Sequence describes ordered children. Empty and singleton sequences reduce to
// their identity values; larger sequences preserve order and nesting.
func Sequence(plans ...Plan) Plan {
	switch len(plans) {
	case 0:
		return Plan{}
	case 1:
		return plans[0]
	default:
		return Plan{kind: KindSequence, children: slices.Clone(plans)}
	}
}

// Transaction describes a conceptual transactional region. The node is always
// preserved because the boundary itself carries information.
func Transaction(plans ...Plan) Plan {
	return Plan{kind: KindTransaction, children: slices.Clone(plans)}
}

// Kind returns the structural form of p.
func (p Plan) Kind() Kind { return p.kind }

// Name returns a Step name and the empty string for every other Kind.
func (p Plan) Name() string { return p.name }

// Children returns an owned copy of p's children, or nil when p has none.
func (p Plan) Children() []Plan { return slices.Clone(p.children) }
