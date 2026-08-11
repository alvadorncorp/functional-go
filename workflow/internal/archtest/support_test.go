package archtest

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	workflowModule     = "github.com/alvadorncorp/gopackages/workflow"
	actionPackage      = workflowModule + "/action"
	transactionPackage = workflowModule + "/transaction"
	archtestPackage    = workflowModule + "/internal/archtest"
	// examples live under the monorepo root (examples/workflow/<slug>), not
	// nested inside this module. The prefix filters any accidental dependency.
	examplesModulePrefix = "github.com/alvadorncorp/gopackages/examples/workflow"
)

// listedPackage is the subset of `go list -json` this gate reads.
type listedPackage struct {
	Dir        string
	GoFiles    []string
	ImportPath string
	Imports    []string
	Standard   bool
}

// forbiddenDirectImports names the standard-library packages whose presence in
// a production file would contradict a documented non-goal. context is the only
// runtime concept this module accepts; sync, sync/atomic, and time would signal
// memoization, parallelism, retry, or timeout creeping into v0.1.
var forbiddenDirectImports = map[string]struct{}{
	"database/sql":  {},
	"encoding/json": {},
	"io/fs":         {},
	"log/slog":      {},
	"net/http":      {},
	"os":            {},
	"os/exec":       {},
	"reflect":       {},
	"sync":          {},
	"sync/atomic":   {},
	"time":          {},
	"unsafe":        {},
}

// forbiddenExportedNames extends the modeling ban list with the names this
// module's non-goals rule out. A symbol appearing here means an engine, a
// container, a registry, a resilience policy, or a distributed-transaction
// promise entered a module that guarantees none of them.
var forbiddenExportedNames = map[string]struct{}{
	"Await":                  {},
	"Boot":                   {},
	"CommandBus":             {},
	"CommandDispatcher":      {},
	"CommandRegistry":        {},
	"Container":              {},
	"DistributedTransaction": {},
	"Discover":               {},
	"Dispatch":               {},
	"Engine":                 {},
	"EventBus":               {},
	"EventSourcedRepository": {},
	"EventStore":             {},
	"EventStream":            {},
	"Fork":                   {},
	"Future":                 {},
	"GlobalTransaction":      {},
	"HTTPHandler":            {},
	"HandlerRegistry":        {},
	"Inject":                 {},
	"Join":                   {},
	"Memoize":                {},
	"Outbox":                 {},
	"Parallel":               {},
	"ParZip":                 {},
	"Promise":                {},
	"Race":                   {},
	"Register":               {},
	"Registry":               {},
	"Repository":             {},
	"Resolve":                {},
	"Retry":                  {},
	"Runner":                 {},
	"Runtime":                {},
	"Saga":                   {},
	"SagaEngine":             {},
	"Scan":                   {},
	"Scheduler":              {},
	"SnapshotStore":          {},
	"Timeout":                {},
	"Worker":                 {},
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

func goModPath(t *testing.T) string {
	t.Helper()

	output, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	return strings.TrimSpace(string(output))
}

// productionPackages lists this module's production packages.
//
// Packages with no GoFiles (the gate package itself is test-only) are skipped.
// Examples live under github.com/alvadorncorp/gopackages/examples/workflow/...
// and no longer share this module's import-path prefix.
func productionPackages(t *testing.T) []listedPackage {
	t.Helper()

	packages := goList(t, "-json", workflowModule+"/...")
	production := make([]listedPackage, 0, len(packages))
	for i := range packages {
		if isExamples(packages[i].ImportPath) ||
			packages[i].ImportPath == archtestPackage ||
			len(packages[i].GoFiles) == 0 {
			continue
		}
		production = append(production, packages[i])
	}
	return production
}

func isExamples(importPath string) bool {
	return importPath == examplesModulePrefix ||
		strings.HasPrefix(importPath, examplesModulePrefix+"/")
}

func dependencyAllowed(pkg *listedPackage) bool {
	return pkg.Standard || pkg.ImportPath == workflowModule ||
		strings.HasPrefix(pkg.ImportPath, workflowModule+"/")
}

func importForbidden(importPath string) bool {
	_, forbidden := forbiddenDirectImports[importPath]
	return forbidden
}

func productionFiles(t *testing.T, pkg *listedPackage) []*ast.File {
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

func productionNames(t *testing.T, pkg *listedPackage) []string {
	t.Helper()

	var names []string
	for _, file := range productionFiles(t, pkg) {
		for _, declaration := range file.Decls {
			names = append(names, exportedNames(declaration)...)
		}
	}
	return names
}

func exportedNames(declaration ast.Decl) []string {
	var names []string
	switch declaration := declaration.(type) {
	case *ast.FuncDecl:
		if declaration.Recv == nil && ast.IsExported(declaration.Name.Name) {
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
	return names
}
