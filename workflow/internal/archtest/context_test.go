package archtest

import (
	"go/ast"
	"testing"
)

// contextValueSelectors names the calls that would smuggle a value — a
// transaction handle above all — through a context.Context.
func contextValueSelectors() map[string]struct{} {
	return map[string]struct{}{
		"WithValue": {},
		"Value":     {},
	}
}

// TestNoContextValueStorage proves the module neither writes nor reads context
// values in production. This is the mechanical form of the architectural rule
// that a context carries lifecycle — cancellation and deadlines — while
// capabilities travel as explicit arguments (R10).
func TestNoContextValueStorage(t *testing.T) {
	selectors := contextValueSelectors()

	for _, production := range productionPackages(t) {
		for _, file := range productionFiles(t, &production) {
			ast.Inspect(file, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if _, watched := selectors[selector.Sel.Name]; !watched {
					return true
				}
				receiver, ok := selector.X.(*ast.Ident)
				if ok && (receiver.Name == "context" || receiver.Name == "ctx") {
					t.Errorf("%s uses %s.%s; capabilities must travel as arguments",
						production.ImportPath, receiver.Name, selector.Sel.Name)
				}
				return true
			})
		}
	}
}
