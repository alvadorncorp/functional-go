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

func TestReplayRangesDirectlyOverSequence(t *testing.T) {
	pkg := reducerProductionPackage(t)
	path := filepath.Join(pkg.Dir, "replay.go")
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
			t.Errorf("Replay imports functional package %s", importPath)
		}
		if strings.HasPrefix(importPath, functionalModule+"/primitives/fold") {
			t.Errorf("Replay imports materializing fold package %s", importPath)
		}
	}

	var replay *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "Replay" {
			replay = function
			break
		}
	}
	if replay == nil {
		t.Fatal("replay.go has no Replay function")
	}

	var eventsRange *ast.RangeStmt
	ast.Inspect(replay.Body, func(node ast.Node) bool {
		rangeStatement, ok := node.(*ast.RangeStmt)
		if !ok {
			return true
		}
		ident, ok := rangeStatement.X.(*ast.Ident)
		if !ok || ident.Name != "events" {
			return true
		}
		if eventsRange != nil {
			t.Fatal("Replay has more than one range over events")
		}
		eventsRange = rangeStatement
		return true
	})
	if eventsRange == nil {
		t.Fatal("Replay does not range directly over events")
	}

	event, ok := eventsRange.Key.(*ast.Ident)
	if !ok || event.Name != "event" || eventsRange.Value != nil {
		t.Fatal("Replay range must bind exactly one event variable")
	}
	if len(eventsRange.Body.List) != 1 {
		t.Fatalf("Replay range statements = %d, want 1", len(eventsRange.Body.List))
	}
	assignment, ok := eventsRange.Body.List[0].(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN ||
		len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		t.Fatal("Replay range must assign one reduced result")
	}
	state, lhsOK := assignment.Lhs[0].(*ast.Ident)
	call, rhsOK := assignment.Rhs[0].(*ast.CallExpr)
	if !lhsOK || state.Name != "state" || !rhsOK {
		t.Fatal("Replay range must assign state from a call")
	}
	callee, callOK := call.Fun.(*ast.Ident)
	if !callOK || callee.Name != "reduce" || len(call.Args) != 2 {
		t.Fatal("Replay range must assign state = reduce(state, event)")
	}
	current, currentOK := call.Args[0].(*ast.Ident)
	input, inputOK := call.Args[1].(*ast.Ident)
	if !currentOK || current.Name != "state" ||
		!inputOK || input.Name != event.Name {
		t.Fatal("Replay range threads state and the ranged event")
	}
}

func TestFunctionalTypeDetectionSelfCheck(t *testing.T) {
	functionalExpression, err := parser.ParseExpr("immutable.List[int]")
	if err != nil {
		t.Fatalf("parse functional type: %v", err)
	}
	if !functionalTypeInExpr(functionalExpression, map[string]struct{}{"immutable": {}}, nil) {
		t.Fatal("public-signature gate accepted a functional type")
	}
	stdlibExpression, err := parser.ParseExpr("iter.Seq[int]")
	if err != nil {
		t.Fatalf("parse standard-library type: %v", err)
	}
	if functionalTypeInExpr(stdlibExpression, map[string]struct{}{"immutable": {}}, nil) {
		t.Fatal("public-signature gate rejected a standard-library type")
	}
	aliasFile, err := parser.ParseFile(token.NewFileSet(), "synthetic.go", `package synthetic
import "github.com/alvadorncorp/gopackages/functional/immutable"
type eventList = immutable.List[int]
func Events() eventList { panic("unreachable") }
`, 0)
	if err != nil {
		t.Fatalf("parse alias fixture: %v", err)
	}
	aliasOrigins := functionalTypeOrigins([]*ast.File{aliasFile})
	aliasImports, _ := functionalImportAliases(aliasFile)
	aliasRejected := false
	for _, declaration := range aliasFile.Decls {
		for _, expression := range exportedTypeExpressions(declaration) {
			aliasRejected = aliasRejected || functionalTypeInExpr(expression, aliasImports, aliasOrigins)
		}
	}
	if !aliasRejected {
		t.Fatal("public-signature gate accepted a local alias to a functional type")
	}
}
