// Package plan provides immutable, non-executable descriptions of workflow
// structure. A Plan is application-declared data: it never receives a context,
// runs an action, opens a transaction, or performs I/O.
package plan

import (
	"slices"
	"strings"
)

// Kind identifies the structural form of a Plan.
type Kind uint8

const (
	KindEmpty Kind = iota
	KindStep
	KindSequence
	KindTransaction
)

// String returns the lowercase kind name, or "unknown" for an undefined Kind.
func (k Kind) String() string {
	switch k {
	case KindEmpty:
		return "empty"
	case KindStep:
		return "step"
	case KindSequence:
		return "sequence"
	case KindTransaction:
		return "transaction"
	default:
		return "unknown"
	}
}

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

// String returns a canonical description of p's declared structure.
// Step names appear verbatim, without escaping or a round-trip parsing promise.
func (p Plan) String() string {
	switch p.kind {
	case KindSequence, KindTransaction:
		var b strings.Builder
		writePlan(&b, &p)
		return b.String()
	case KindStep:
		return p.name
	default:
		return p.kind.String()
	}
}

func writePlan(b *strings.Builder, p *Plan) {
	switch p.kind {
	case KindSequence, KindTransaction:
		b.WriteString(p.kind.String())
		b.WriteByte('(')
		for i := range p.children {
			if i > 0 {
				b.WriteString(", ")
			}
			writePlan(b, &p.children[i])
		}
		b.WriteByte(')')
	case KindStep:
		b.WriteString(p.name)
	default:
		b.WriteString(p.kind.String())
	}
}
