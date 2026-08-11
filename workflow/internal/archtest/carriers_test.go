package archtest

import (
	"go/ast"
	"reflect"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func TestPairIsAFieldOnlyCarrier(t *testing.T) {
	pair := reflect.TypeFor[action.Pair[int, string]]()
	if pair.NumField() != 2 || pair.Field(0).Name != "First" || pair.Field(1).Name != "Second" {
		t.Fatalf("Pair fields = %v, want exactly First and Second", pair)
	}
	if pair.NumMethod() != 0 {
		t.Fatalf("Pair methods = %d, want 0", pair.NumMethod())
	}
	assertNoCarrierMethods(t, "Pair")
}

func assertNoCarrierMethods(t *testing.T, carrier string) {
	t.Helper()
	for _, production := range productionPackages(t) {
		if production.ImportPath != actionPackage {
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
