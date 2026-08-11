package archtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSurfaceIsExercised proves every exported core symbol is used by something
// other than its own package's tests — a sibling core package, a doc example, or
// a monorepo examples/workflow slug.
//
// This is the simplicity criterion made mechanical: an API kept only for
// symmetry has no consumer and should not ship.
func TestSurfaceIsExercised(t *testing.T) {
	consumers := consumerSources(t)

	for pkg, names := range exactSurface() {
		short := shortPackageName(pkg)
		for _, name := range names {
			qualified := short + "." + name
			if !strings.Contains(consumers, qualified) {
				t.Errorf("%s is exported but never used outside its own package; "+
					"either exercise it in an example or remove it from the surface", qualified)
			}
		}
	}
}

func shortPackageName(importPath string) string {
	if importPath == workflowModule {
		return "workflow"
	}
	return importPath[strings.LastIndex(importPath, "/")+1:]
}

// consumerSources concatenates every Go file that legitimately counts as a
// consumer of the core surface: this module's example_test.go files, and every
// file under monorepo examples/workflow/.
//
// A package's own behaviour tests are deliberately excluded. A symbol used only
// by the tests written for it has no consumer, and counting those tests would
// make this gate pass vacuously — which is exactly the failure mode it exists to
// prevent.
func consumerSources(t *testing.T) string {
	t.Helper()

	root := moduleRoot(t)
	workspaceRoot := filepath.Dir(root)
	examplesRoot := filepath.Join(workspaceRoot, "examples", "workflow")
	separator := string(filepath.Separator)
	archtestDir := separator + "internal" + separator + "archtest" + separator
	examplesDir := separator + "examples" + separator

	var builder strings.Builder
	walk := func(path string, entry os.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir(), !strings.HasSuffix(path, ".go"):
			return nil
		case strings.Contains(path, archtestDir):
			return nil
		case strings.HasSuffix(path, "_test.go") &&
			!strings.HasSuffix(path, "example_test.go") &&
			!strings.Contains(path, examplesDir):
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		builder.Write(data)
		return nil
	}

	if err := filepath.WalkDir(root, walk); err != nil {
		t.Fatalf("scan consumers in %s: %v", root, err)
	}
	if err := filepath.WalkDir(examplesRoot, walk); err != nil {
		t.Fatalf("scan consumers in %s: %v", examplesRoot, err)
	}

	// go vet keeps the walk honest: a consumer that does not compile is not a
	// consumer. It must run from the module root, not from this package.
	vet := exec.Command("go", "vet", "./...")
	vet.Dir = root
	if output, err := vet.CombinedOutput(); err != nil {
		t.Fatalf("go vet: %v\n%s", err, output)
	}

	return builder.String()
}
