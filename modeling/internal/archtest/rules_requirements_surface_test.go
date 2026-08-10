package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestRulesAndRequirementsHaveExactPublicSurface(t *testing.T) {
	tests := map[string][]string{
		"requirement.go": {"AllRequirements", "AnyRequirement", "Requirement"},
		"rule.go":        {"AllRules", "AnyRule", "Rule"},
	}
	for filename, want := range tests {
		t.Run(filename, func(t *testing.T) {
			got := exportedNamesInSource(t, filepath.Join("..", "..", filename))
			if !slices.Equal(got, want) {
				t.Fatalf("exports in %s = %v, want %v", filename, got, want)
			}
		})
	}
}

func exportedNamesInSource(t *testing.T, filename string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}

	var names []string
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.GenDecl:
			for _, specification := range declaration.Specs {
				if typeSpec, ok := specification.(*ast.TypeSpec); ok && ast.IsExported(typeSpec.Name.Name) {
					names = append(names, typeSpec.Name.Name)
				}
			}
		case *ast.FuncDecl:
			if ast.IsExported(declaration.Name.Name) {
				names = append(names, declaration.Name.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}
