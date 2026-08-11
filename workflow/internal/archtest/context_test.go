package archtest

import (
	"go/ast"
	"testing"
)

func forbiddenContextSelectors() map[string]struct{} {
	return map[string]struct{}{
		"AfterFunc":         {},
		"Background":        {},
		"TODO":              {},
		"Value":             {},
		"WithCancel":        {},
		"WithCancelCause":   {},
		"WithDeadline":      {},
		"WithDeadlineCause": {},
		"WithoutCancel":     {},
		"WithTimeout":       {},
		"WithTimeoutCause":  {},
		"WithValue":         {},
	}
}

func forbiddenContextUse(file *ast.File) (string, bool) {
	aliases := map[string]struct{}{}
	for _, imported := range file.Imports {
		if imported.Path.Value != `"context"` {
			continue
		}
		alias := "context"
		if imported.Name != nil {
			alias = imported.Name.Name
		}
		if alias == "." {
			return "dot-import of context", true
		}
		aliases[alias] = struct{}{}
	}

	selectors := forbiddenContextSelectors()
	var violation string
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if selector.Sel.Name == "Value" {
			violation = "context value access via .Value"
			return false
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, contextAlias := aliases[receiver.Name]; !contextAlias {
			return true
		}
		if _, watched := selectors[selector.Sel.Name]; watched {
			violation = receiver.Name + "." + selector.Sel.Name
			return false
		}
		return true
	})
	return violation, violation != ""
}

// TestNoHiddenContextPolicy proves production neither stores capabilities in a
// context nor creates/shields a context behind a combinator. Context lifecycle
// policy remains explicit at the caller or adapter.
func TestNoHiddenContextPolicy(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, file := range productionFiles(t, &production) {
			if violation, forbidden := forbiddenContextUse(file); forbidden {
				t.Errorf("%s uses %s; context policy must remain explicit",
					production.ImportPath, violation)
			}
		}
	}
}
