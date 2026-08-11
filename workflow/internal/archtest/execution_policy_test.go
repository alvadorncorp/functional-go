package archtest

import (
	"go/ast"
	"testing"
)

func TestNoGoStatementsOrRecoverCalls(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, file := range productionFiles(t, &production) {
			ast.Inspect(file, func(node ast.Node) bool {
				if violation := forbiddenExecutionNode(node); violation != "" {
					t.Errorf("%s contains %s", production.ImportPath, violation)
				}
				return true
			})
		}
	}
}

func forbiddenExecutionNode(node ast.Node) string {
	switch node := node.(type) {
	case *ast.GoStmt:
		return "a go statement; workflow execution must be sequential"
	case *ast.CallExpr:
		identifier, ok := node.Fun.(*ast.Ident)
		if ok && identifier.Name == "recover" {
			return "a recover call; panics must propagate"
		}
	}
	return ""
}
