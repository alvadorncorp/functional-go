package archtest

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

var expectedValuesAPI = map[string][]string{
	pkgValuesOptional: {
		"Equal", "Filter", "FlatMap", "From", "Get", "Inspect", "IsNone",
		"IsSome", "Map", "MustGet", "None", "Optional", "OrElse", "OrZero", "Some",
	},
	pkgValuesResult: {
		"Equal", "Err", "ErrValue", "FlatMap", "Get", "IsErr", "IsOk", "Map",
		"MapErr", "MustGet", "Ok", "OrElse", "OrZero", "Result",
	},
	pkgValuesEither: {
		"Bimap", "Either", "FlatMapRight", "IsLeft", "IsRight", "Left",
		"LeftValue", "MapLeft", "MapRight", "Match", "Right", "RightValue",
	},
	pkgValuesNonEmpty: {
		"Append", "At", "FromSlice", "Head", "Items", "Last", "Len", "Map",
		"NonEmpty", "Of", "Tail",
	},
	pkgValuesValidated: {
		"Errors", "Get", "Invalid", "InvalidNonEmpty", "IsInvalid", "IsValid",
		"Map", "Map2", "MapErrors", "Valid", "Validated",
	},
}

func TestValuesPublicAPI(t *testing.T) {
	t.Parallel()
	for _, pkg := range valuesPackages {
		pkg := pkg
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			got := valuesExportedNames(t, pkg)
			want := slices.Clone(expectedValuesAPI[pkg])
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("exported API for %s = %v, want %v", pkg, got, want)
			}
		})
	}
}

func valuesExportedNames(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	if err != nil {
		t.Fatalf("go list -json %s: %v", pkg, err)
	}
	var listed struct {
		Dir     string
		GoFiles []string
	}
	if err := json.Unmarshal(out, &listed); err != nil {
		t.Fatalf("decode go list -json %s: %v", pkg, err)
	}

	set := token.NewFileSet()
	var names []string
	for _, name := range listed.GoFiles {
		path := filepath.Join(listed.Dir, name)
		file, err := parser.ParseFile(set, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Name.IsExported() {
					names = append(names, declaration.Name.Name)
				}
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					switch specification := specification.(type) {
					case *ast.TypeSpec:
						if specification.Name.IsExported() {
							names = append(names, specification.Name.Name)
						}
					case *ast.ValueSpec:
						for _, valueName := range specification.Names {
							if valueName.IsExported() {
								names = append(names, valueName.Name)
							}
						}
					}
				}
			}
		}
	}
	slices.Sort(names)
	return names
}
