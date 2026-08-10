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

func TestReplayUsesValuesWithoutMaterialization(t *testing.T) {
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

	valuesCalls := 0
	itemsCalls := 0
	ast.Inspect(replay.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch selector.Sel.Name {
		case "Values":
			receiver, ok := selector.X.(*ast.Ident)
			if ok && receiver.Name == "events" {
				valuesCalls++
			}
		case "Items":
			itemsCalls++
		}
		return true
	})

	if valuesCalls != 1 {
		t.Errorf("Replay Values calls = %d, want exactly 1", valuesCalls)
	}
	if itemsCalls != 0 {
		t.Errorf("Replay Items calls = %d, want 0", itemsCalls)
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
