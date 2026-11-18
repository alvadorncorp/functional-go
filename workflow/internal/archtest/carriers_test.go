package archtest

import (
	"go/ast"
	"reflect"
	"testing"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/resilience"
)

func TestPairIsAFieldOnlyCarrier(t *testing.T) {
	pair := reflect.TypeFor[action.Pair[int, string]]()
	if pair.NumField() != 2 || pair.Field(0).Name != "First" || pair.Field(1).Name != "Second" {
		t.Fatalf("Pair fields = %v, want exactly First and Second", pair)
	}
	if pair.NumMethod() != 0 {
		t.Fatalf("Pair methods = %d, want 0", pair.NumMethod())
	}
	assertNoCarrierMethods(t, actionPackage, "Pair")
}

func TestAttemptResultIsAFieldOnlyCarrier(t *testing.T) {
	result := reflect.TypeFor[action.AttemptResult[int]]()
	if result.NumField() != 2 || result.Field(0).Name != "Value" || result.Field(1).Name != "Err" {
		t.Fatalf("AttemptResult fields = %v, want exactly Value and Err", result)
	}
	if result.NumMethod() != 0 {
		t.Fatalf("AttemptResult methods = %d, want 0", result.NumMethod())
	}
	assertNoCarrierMethods(t, actionPackage, "AttemptResult")
}

func TestRetryPolicyIsAFieldOnlyCarrier(t *testing.T) {
	policy := reflect.TypeFor[resilience.RetryPolicy]()
	if policy.NumField() != 3 {
		t.Fatalf("RetryPolicy fields = %d, want 3", policy.NumField())
	}
	if policy.Field(0).Name != "Attempts" || policy.Field(0).Type.Kind() != reflect.Int {
		t.Fatalf("RetryPolicy field 0 = %s %s, want Attempts int", policy.Field(0).Name, policy.Field(0).Type)
	}
	when := policy.Field(1)
	if when.Name != "When" || when.Type.Kind() != reflect.Func ||
		when.Type.NumIn() != 1 || when.Type.NumOut() != 1 ||
		when.Type.In(0) != reflect.TypeFor[error]() || when.Type.Out(0).Kind() != reflect.Bool {
		t.Fatalf("RetryPolicy field 1 = %s %s, want When func(error) bool", when.Name, when.Type)
	}
	backoff := policy.Field(2)
	if backoff.Name != "Backoff" || backoff.Type != reflect.TypeFor[resilience.Backoff]() {
		t.Fatalf("RetryPolicy field 2 = %s %s, want Backoff resilience.Backoff", backoff.Name, backoff.Type)
	}
	if policy.NumMethod() != 0 {
		t.Fatalf("RetryPolicy methods = %d, want 0", policy.NumMethod())
	}
	assertNoCarrierMethods(t, resiliencePackage, "RetryPolicy")
}

func assertNoCarrierMethods(t *testing.T, pkg, carrier string) {
	t.Helper()
	for _, production := range productionPackages(t) {
		if production.ImportPath != pkg {
			continue
		}
		for _, file := range productionFiles(t, &production) {
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok || function.Recv == nil || len(function.Recv.List) == 0 {
					continue
				}
				if receiverBaseName(function.Recv.List[0].Type) == carrier {
					t.Errorf("%s declares method %s, want zero methods", carrier, function.Name.Name)
				}
			}
		}
	}
}

func receiverBaseName(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.StarExpr:
		return receiverBaseName(expression.X)
	case *ast.IndexExpr:
		return receiverBaseName(expression.X)
	case *ast.IndexListExpr:
		return receiverBaseName(expression.X)
	default:
		return ""
	}
}
