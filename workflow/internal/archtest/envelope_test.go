package archtest

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const maxProductionFileLines = 500

// expectedFiles pins each production package's file set. A new file is a
// deliberate structural change and must appear here in the same commit.
func expectedFiles() map[string][]string {
	return map[string][]string{
		workflowModule: {"compose.go", "doc.go", "input.go", "panic.go", "workflow.go"},
		actionPackage: {
			"action.go", "attempt.go", "combine.go", "construct.go", "control.go",
			"observe.go", "panic.go", "resource.go", "transform.go",
		},
		transactionPackage: {"transaction.go"},
	}
}

// TestFileEnvelope proves the delivered file set is exactly the decided one and
// that no production file exceeds the project's 500-line ceiling (R11).
func TestFileEnvelope(t *testing.T) {
	envelope := expectedFiles()

	for _, production := range productionPackages(t) {
		want, declared := envelope[production.ImportPath]
		if !declared {
			t.Errorf("%s is not declared in expectedFiles", production.ImportPath)
			continue
		}

		got := slices.Clone(production.GoFiles)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("%s holds %v, want %v", production.ImportPath, got, want)
		}

		for _, name := range production.GoFiles {
			assertFileWithinLineLimit(t, filepath.Join(production.Dir, name))
		}
	}
}

func assertFileWithinLineLimit(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if lines := bytes.Count(data, []byte("\n")); lines > maxProductionFileLines {
		t.Errorf("%s has %d lines, want at most %d", path, lines, maxProductionFileLines)
	}
}
