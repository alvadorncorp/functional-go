package plan_test

import (
	"sync"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/plan"
)

func TestZeroPlanIsEmpty(t *testing.T) {
	var subject plan.Plan
	if subject.Kind() != plan.KindEmpty || subject.Name() != "" || subject.Children() != nil {
		t.Fatalf("zero Plan = kind %v name %q children %v", subject.Kind(), subject.Name(), subject.Children())
	}
}

func TestStep(t *testing.T) {
	subject := plan.Step("load-customer")
	if subject.Kind() != plan.KindStep || subject.Name() != "load-customer" || subject.Children() != nil {
		t.Fatalf("Step = kind %v name %q children %v", subject.Kind(), subject.Name(), subject.Children())
	}
	if plan.Step(" ").Name() != " " {
		t.Fatal("Step normalized a non-empty name")
	}
	assertPanic(t, "workflow/plan: empty step name", func() { plan.Step("") })
}

func TestSequenceIdentityAndOrder(t *testing.T) {
	var empty plan.Plan
	if got := plan.Sequence(); got.Kind() != plan.KindEmpty {
		t.Fatalf("Sequence() kind = %v, want KindEmpty", got.Kind())
	}
	if got := plan.Sequence(empty); got.Kind() != plan.KindEmpty {
		t.Fatalf("Sequence(empty) kind = %v, want KindEmpty", got.Kind())
	}
	a, b, c := plan.Step("a"), plan.Step("b"), plan.Step("c")
	if got := plan.Sequence(a); got.Kind() != plan.KindStep || got.Name() != "a" {
		t.Fatalf("Sequence(a) = %#v, want a", got)
	}
	subject := plan.Sequence(a, b, c)
	children := subject.Children()
	if subject.Kind() != plan.KindSequence || subject.Name() != "" || len(children) != 3 {
		t.Fatalf("Sequence = kind %v name %q children %v", subject.Kind(), subject.Name(), children)
	}
	for i, want := range []string{"a", "b", "c"} {
		if children[i].Name() != want {
			t.Fatalf("child %d = %q, want %q", i, children[i].Name(), want)
		}
	}
}

func TestSequencePreservesNesting(t *testing.T) {
	nested := plan.Sequence(plan.Step("b"), plan.Step("c"))
	subject := plan.Sequence(plan.Step("a"), nested)
	children := subject.Children()
	if len(children) != 2 || children[1].Kind() != plan.KindSequence {
		t.Fatalf("nested Sequence flattened: %#v", children)
	}
}

func TestTransactionAlwaysPreservesBoundary(t *testing.T) {
	empty := plan.Transaction()
	if empty.Kind() != plan.KindTransaction || empty.Name() != "" || empty.Children() != nil {
		t.Fatalf("Transaction() = kind %v name %q children %v", empty.Kind(), empty.Name(), empty.Children())
	}

	one := plan.Transaction(plan.Step("save"))
	oneChildren := one.Children()
	if one.Kind() != plan.KindTransaction || one.Name() != "" || len(oneChildren) != 1 || oneChildren[0].Name() != "save" {
		t.Fatalf("Transaction(one) = kind %v name %q children %v", one.Kind(), one.Name(), oneChildren)
	}

	two := plan.Transaction(plan.Step("save"), plan.Step("outbox"))
	twoChildren := two.Children()
	if two.Kind() != plan.KindTransaction || two.Name() != "" || len(twoChildren) != 2 {
		t.Fatalf("Transaction(two) = kind %v name %q children %v", two.Kind(), two.Name(), twoChildren)
	}
	for i, want := range []string{"save", "outbox"} {
		if twoChildren[i].Name() != want {
			t.Fatalf("Transaction child %d = %q, want %q", i, twoChildren[i].Name(), want)
		}
	}
}

func TestConstructorsAndChildrenOwnTheirSlices(t *testing.T) {
	a, b, changed := plan.Step("a"), plan.Step("b"), plan.Step("changed")

	seqInput := []plan.Plan{a, b}
	seq := plan.Sequence(seqInput...)
	seqInput[0] = changed
	seqFirst := seq.Children()
	seqFirst[1] = changed
	seqSecond := seq.Children()
	if seqSecond[0].Name() != "a" || seqSecond[1].Name() != "b" {
		t.Fatalf("Sequence leaked slice ownership: %v", seqSecond)
	}

	txInput := []plan.Plan{a, b}
	tx := plan.Transaction(txInput...)
	txInput[0] = changed
	txFirst := tx.Children()
	txFirst[1] = changed
	txSecond := tx.Children()
	if txSecond[0].Name() != "a" || txSecond[1].Name() != "b" {
		t.Fatalf("Transaction leaked slice ownership: %v", txSecond)
	}
}

func TestKindString(t *testing.T) {
	tests := []struct {
		kind plan.Kind
		want string
	}{
		{plan.KindEmpty, "empty"},
		{plan.KindStep, "step"},
		{plan.KindSequence, "sequence"},
		{plan.KindTransaction, "transaction"},
		{plan.Kind(4), "unknown"},
		{plan.Kind(255), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Fatalf("Kind(%d).String() = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

func TestPlanString(t *testing.T) {
	var empty plan.Plan
	tests := []struct {
		name    string
		subject plan.Plan
		want    string
	}{
		{"zero", empty, "empty"},
		{"step", plan.Step("a"), "a"},
		{"verbatim name", plan.Step("  ação(a, b)\n "), "  ação(a, b)\n "},
		{"empty sequence", plan.Sequence(), "empty"},
		{"singleton sequence", plan.Sequence(plan.Step("a")), "a"},
		{"empty transaction", plan.Transaction(), "transaction()"},
		{"singleton transaction", plan.Transaction(plan.Step("a")), "transaction(a)"},
		{
			"canonical",
			plan.Sequence(plan.Step("a"), plan.Step("b"), plan.Transaction(plan.Step("c"), plan.Step("d"))),
			"sequence(a, b, transaction(c, d))",
		},
		{
			"nested sequence",
			plan.Sequence(plan.Step("a"), plan.Sequence(plan.Step("b"), plan.Step("c"))),
			"sequence(a, sequence(b, c))",
		},
		{
			"empty children and boundaries",
			plan.Sequence(empty, plan.Transaction(), plan.Transaction(plan.Transaction())),
			"sequence(empty, transaction(), transaction(transaction()))",
		},
		{
			"nested verbatim name",
			plan.Transaction(plan.Step("  ação(a, b)\n ")),
			"transaction(  ação(a, b)\n )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 2 {
				if got := tt.subject.String(); got != tt.want {
					t.Fatalf("String() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestPlanStringPreservesSnapshots(t *testing.T) {
	subject := plan.Sequence(plan.Step("a"), plan.Transaction(plan.Step("b")))
	snapshot := subject.Children()
	if got := subject.String(); got != "sequence(a, transaction(b))" {
		t.Fatalf("String() = %q", got)
	}
	for _, children := range [][]plan.Plan{snapshot, subject.Children()} {
		if len(children) != 2 || children[0].Name() != "a" || children[1].Kind() != plan.KindTransaction {
			t.Fatalf("String mutated children: %#v", children)
		}
		nested := children[1].Children()
		if len(nested) != 1 || nested[0].Name() != "b" {
			t.Fatalf("String mutated nested children: %#v", nested)
		}
	}
}

func TestPlanSupportsConcurrentReads(t *testing.T) {
	subject := plan.Sequence(plan.Step("a"), plan.Transaction(plan.Step("b")))
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				if subject.Kind() != plan.KindSequence || len(subject.Children()) != 2 {
					t.Errorf("concurrent read observed invalid Plan")
					return
				}
			}
		}()
	}
	wait.Wait()
}

func assertPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	fn()
}
