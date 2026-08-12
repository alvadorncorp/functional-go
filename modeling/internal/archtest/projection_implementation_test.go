package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProjectRangesDirectlyOverSequence(t *testing.T) {
	pkg := projectionProductionPackage(t)
	path := filepath.Join(pkg.Dir, "project.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	for _, imported := range file.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			t.Fatalf("unquote import %s: %v", imported.Path.Value, err)
		}
		if strings.HasPrefix(importPath, functionalModule) {
			t.Errorf("Project imports functional package %s", importPath)
		}
		if strings.HasPrefix(importPath, functionalModule+"/primitives/fold") {
			t.Errorf("Project imports materializing fold package %s", importPath)
		}
		if importPath == reducerPackage || strings.HasPrefix(importPath, reducerPackage+"/") {
			t.Errorf("Project imports semantic sibling %s", importPath)
		}
	}

	var project *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "Project" {
			project = function
			break
		}
	}
	if project == nil {
		t.Fatal("project.go has no Project function")
	}

	var eventsRange *ast.RangeStmt
	ast.Inspect(project.Body, func(node ast.Node) bool {
		rangeStatement, ok := node.(*ast.RangeStmt)
		if !ok {
			return true
		}
		ident, ok := rangeStatement.X.(*ast.Ident)
		if !ok || ident.Name != "events" {
			return true
		}
		if eventsRange != nil {
			t.Fatal("Project has more than one range over events")
		}
		eventsRange = rangeStatement
		return true
	})
	if eventsRange == nil {
		t.Fatal("Project does not range directly over events")
	}

	event, ok := eventsRange.Key.(*ast.Ident)
	if !ok || event.Name != "event" || eventsRange.Value != nil {
		t.Fatal("Project range must bind exactly one event variable")
	}
	if len(eventsRange.Body.List) != 1 {
		t.Fatalf("Project range statements = %d, want 1", len(eventsRange.Body.List))
	}
	assignment, ok := eventsRange.Body.List[0].(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN ||
		len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		t.Fatal("Project range must assign one projected result")
	}
	projected, lhsOK := assignment.Lhs[0].(*ast.Ident)
	call, rhsOK := assignment.Rhs[0].(*ast.CallExpr)
	if !lhsOK || projected.Name != "projected" || !rhsOK {
		t.Fatal("Project range must assign projected from a call")
	}
	callee, callOK := call.Fun.(*ast.Ident)
	if !callOK || callee.Name != "project" || len(call.Args) != 2 {
		t.Fatal("Project range must assign projected = project(projected, event)")
	}
	current, currentOK := call.Args[0].(*ast.Ident)
	input, inputOK := call.Args[1].(*ast.Ident)
	if !currentOK || current.Name != "projected" ||
		!inputOK || input.Name != event.Name {
		t.Fatal("Project range threads projected and the ranged event")
	}
}
