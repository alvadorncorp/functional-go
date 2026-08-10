package archtest

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	modelingModule      = "github.com/alvadorncorp/gopackages/modeling"
	functionalModule    = "github.com/alvadorncorp/gopackages/functional"
	functionalImmutable = functionalModule + "/immutable"
	archtestPackage     = modelingModule + "/internal/archtest"
)

type listedPackage struct {
	Dir        string
	GoFiles    []string
	ImportPath string
	Imports    []string
	Standard   bool
}

var forbiddenDirectImports = map[string]struct{}{
	"database/sql": {},
	"io/fs":        {},
	"net/http":     {},
	"os":           {},
	"reflect":      {},
	"unsafe":       {},
}

var forbiddenExportedNames = map[string]struct{}{
	"Boot":                   {},
	"CommandBus":             {},
	"CommandDispatcher":      {},
	"CommandRegistry":        {},
	"Discover":               {},
	"Dispatch":               {},
	"EventSourcedRepository": {},
	"EventStore":             {},
	"EventStream":            {},
	"HandlerRegistry":        {},
	"Projection":             {},
	"Register":               {},
	"Replay":                 {},
	"Scan":                   {},
	"SnapshotStore":          {},
}

func goList(t *testing.T, args ...string) []listedPackage {
	t.Helper()
	command := exec.Command("go", append([]string{"list"}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(append([]string{"list"}, args...), " "), err, output)
	}

	decoder := json.NewDecoder(strings.NewReader(string(output)))
	var packages []listedPackage
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode go list output: %v", err)
		}
		packages = append(packages, pkg)
	}
	return packages
}

func productionPackages(t *testing.T) []listedPackage {
	t.Helper()
	packages := goList(t, "-json", modelingModule+"/...")
	production := make([]listedPackage, 0, len(packages))
	for _, pkg := range packages {
		if pkg.ImportPath != archtestPackage {
			production = append(production, pkg)
		}
	}
	return production
}

func dependencyAllowed(pkg listedPackage) bool {
	return pkg.Standard || pkg.ImportPath == modelingModule ||
		strings.HasPrefix(pkg.ImportPath, modelingModule+"/") ||
		pkg.ImportPath == functionalModule ||
		strings.HasPrefix(pkg.ImportPath, functionalModule+"/")
}

func directImportForbidden(importPath string) bool {
	if _, forbidden := forbiddenDirectImports[importPath]; forbidden {
		return true
	}
	if importPath == functionalModule || strings.HasPrefix(importPath, functionalModule+"/") {
		return importPath != functionalImmutable
	}
	return false
}

func TestOnlyApprovedDependencyClosure(t *testing.T) {
	for _, production := range productionPackages(t) {
		t.Run(production.ImportPath, func(t *testing.T) {
			for _, dependency := range goList(t, "-deps", "-json", production.ImportPath) {
				if !dependencyAllowed(dependency) {
					t.Errorf("%s depends on unapproved %s", production.ImportPath, dependency.ImportPath)
				}
			}
		})
	}
}

func TestNoForbiddenDirectImports(t *testing.T) {
	for _, production := range productionPackages(t) {
		t.Run(production.ImportPath, func(t *testing.T) {
			for _, imported := range production.Imports {
				if directImportForbidden(imported) {
					t.Errorf("%s directly imports forbidden %s", production.ImportPath, imported)
				}
			}
		})
	}
}

func TestNoFunctionalTypesInPublicAPI(t *testing.T) {
	for _, production := range productionPackages(t) {
		t.Run(production.ImportPath, func(t *testing.T) {
			files := productionFiles(t, production)
			origins := functionalTypeOrigins(files)
			for _, file := range files {
				aliases, dotImport := functionalImportAliases(file)
				if dotImport {
					t.Error("functional dot imports are forbidden")
					continue
				}
				for _, declaration := range file.Decls {
					for _, expression := range exportedTypeExpressions(declaration) {
						if functionalTypeInExpr(expression, aliases, origins) {
							t.Errorf("%s exports a functional type", production.ImportPath)
						}
					}
				}
			}
		})
	}
}

func TestFunctionalDoesNotDependOnModeling(t *testing.T) {
	for _, dependency := range goList(t, "-deps", "-json", functionalModule+"/...") {
		if dependency.ImportPath == modelingModule || strings.HasPrefix(dependency.ImportPath, modelingModule+"/") {
			t.Errorf("functional must not depend on %s", dependency.ImportPath)
		}
	}
}

func TestNoMagicRuntimeAPI(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, name := range productionNames(t, production) {
			if _, forbidden := forbiddenExportedNames[name]; forbidden {
				t.Errorf("%s exports forbidden runtime API %s", production.ImportPath, name)
			}
		}
	}
}

func TestNoMutableProductionGlobals(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, file := range productionFiles(t, production) {
			for _, declaration := range file.Decls {
				generic, ok := declaration.(*ast.GenDecl)
				if ok && generic.Tok == token.VAR {
					t.Errorf("%s has a package-level var declaration", production.ImportPath)
				}
			}
		}
	}
}

func TestPackageEnumerationIsComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, pkg := range goList(t, "-json", modelingModule+"/...") {
		seen[pkg.ImportPath] = true
	}
	for _, required := range []string{modelingModule, archtestPackage} {
		if !seen[required] {
			t.Errorf("package enumeration is too narrow: missing %s", required)
		}
	}
}

func TestDependencyGateCanFail(t *testing.T) {
	if dependencyAllowed(listedPackage{ImportPath: "example.com/unapproved"}) {
		t.Fatal("dependency gate accepted an external package")
	}
	for _, imported := range []string{
		"reflect",
		"unsafe",
		functionalModule,
		functionalModule + "/internal/core",
		functionalModule + "/values/result",
	} {
		if !directImportForbidden(imported) {
			t.Errorf("direct-import gate accepted %s", imported)
		}
	}
	if directImportForbidden(functionalImmutable) {
		t.Fatalf("direct-import gate rejected approved %s", functionalImmutable)
	}
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
	if _, forbidden := forbiddenExportedNames["Dispatch"]; !forbidden {
		t.Fatal("runtime API gate cannot reject Dispatch")
	}
}

func functionalImportAliases(file *ast.File) (map[string]struct{}, bool) {
	aliases := map[string]struct{}{}
	dotImport := false
	for _, imported := range file.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil || (importPath != functionalModule && !strings.HasPrefix(importPath, functionalModule+"/")) {
			continue
		}
		if imported.Name != nil {
			switch imported.Name.Name {
			case "_":
				continue
			case ".":
				dotImport = true
				continue
			default:
				aliases[imported.Name.Name] = struct{}{}
				continue
			}
		}
		aliases[pathpkg.Base(importPath)] = struct{}{}
	}
	return aliases, dotImport
}

func exportedTypeExpressions(declaration ast.Decl) []ast.Expr {
	switch declaration := declaration.(type) {
	case *ast.FuncDecl:
		if ast.IsExported(declaration.Name.Name) {
			return fieldListTypeExpressions(
				declaration.Type.TypeParams,
				declaration.Type.Params,
				declaration.Type.Results,
			)
		}
	case *ast.GenDecl:
		var expressions []ast.Expr
		for _, specification := range declaration.Specs {
			switch specification := specification.(type) {
			case *ast.TypeSpec:
				if !ast.IsExported(specification.Name.Name) {
					continue
				}
				expressions = append(expressions, typeOriginExpressions(specification)...)
			case *ast.ValueSpec:
				if specification.Type != nil && anyExported(specification.Names) {
					expressions = append(expressions, specification.Type)
				}
			}
		}
		return expressions
	}
	return nil
}

type typeOriginCandidate struct {
	name        string
	expressions []ast.Expr
	aliases     map[string]struct{}
}

func functionalTypeOrigins(files []*ast.File) map[string]struct{} {
	var candidates []typeOriginCandidate
	for _, file := range files {
		aliases, _ := functionalImportAliases(file)
		for _, declaration := range file.Decls {
			generic, ok := declaration.(*ast.GenDecl)
			if !ok || generic.Tok != token.TYPE {
				continue
			}
			for _, specification := range generic.Specs {
				typeSpec := specification.(*ast.TypeSpec)
				candidates = append(candidates, typeOriginCandidate{
					name:        typeSpec.Name.Name,
					expressions: typeOriginExpressions(typeSpec),
					aliases:     aliases,
				})
			}
		}
	}

	origins := map[string]struct{}{}
	for changed := true; changed; {
		changed = false
		for _, candidate := range candidates {
			if _, known := origins[candidate.name]; known {
				continue
			}
			for _, expression := range candidate.expressions {
				if functionalTypeInExpr(expression, candidate.aliases, origins) {
					origins[candidate.name] = struct{}{}
					changed = true
					break
				}
			}
		}
	}
	return origins
}

func typeOriginExpressions(specification *ast.TypeSpec) []ast.Expr {
	expressions := fieldListTypeExpressions(specification.TypeParams)
	switch expression := specification.Type.(type) {
	case *ast.StructType:
		expressions = append(expressions, exportedFieldTypes(expression.Fields.List)...)
	case *ast.InterfaceType:
		expressions = append(expressions, exportedFieldTypes(expression.Methods.List)...)
	default:
		expressions = append(expressions, expression)
	}
	return expressions
}

func fieldListTypeExpressions(lists ...*ast.FieldList) []ast.Expr {
	var expressions []ast.Expr
	for _, list := range lists {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			expressions = append(expressions, field.Type)
		}
	}
	return expressions
}

func exportedFieldTypes(fields []*ast.Field) []ast.Expr {
	var expressions []ast.Expr
	for _, field := range fields {
		if len(field.Names) == 0 || anyExported(field.Names) {
			expressions = append(expressions, field.Type)
		}
	}
	return expressions
}

func anyExported(names []*ast.Ident) bool {
	for _, name := range names {
		if ast.IsExported(name.Name) {
			return true
		}
	}
	return false
}

func functionalTypeInExpr(
	expression ast.Expr,
	aliases map[string]struct{},
	origins map[string]struct{},
) bool {
	ignored := map[*ast.Ident]struct{}{}
	ast.Inspect(expression, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Field:
			for _, name := range node.Names {
				ignored[name] = struct{}{}
			}
		case *ast.SelectorExpr:
			ignored[node.Sel] = struct{}{}
		}
		return true
	})

	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if found {
			return false
		}
		switch node := node.(type) {
		case *ast.SelectorExpr:
			identifier, ok := node.X.(*ast.Ident)
			if ok {
				_, found = aliases[identifier.Name]
			}
		case *ast.Ident:
			if _, skip := ignored[node]; skip {
				break
			}
			_, found = origins[node.Name]
		}
		return !found
	})
	return found
}

func productionFiles(t *testing.T, pkg listedPackage) []*ast.File {
	t.Helper()
	files := make([]*ast.File, 0, len(pkg.GoFiles))
	for _, name := range pkg.GoFiles {
		path := filepath.Join(pkg.Dir, name)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		files = append(files, file)
	}
	return files
}

func productionNames(t *testing.T, pkg listedPackage) []string {
	t.Helper()
	var names []string
	for _, file := range productionFiles(t, pkg) {
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if ast.IsExported(declaration.Name.Name) {
					names = append(names, declaration.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range declaration.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(spec.Name.Name) {
							names = append(names, spec.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								names = append(names, name.Name)
							}
						}
					}
				}
			}
		}
	}
	return names
}
