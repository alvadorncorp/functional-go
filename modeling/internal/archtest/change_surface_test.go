package archtest

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestChangeHasExactProductionFiles(t *testing.T) {
	got := changeFiles(t, false)
	want := []string{"apply.go", "change.go", "composition.go"}
	if !slices.Equal(got, want) {
		t.Fatalf("change production files = %v, want %v", got, want)
	}
}

func TestChangeHasExactTestFiles(t *testing.T) {
	got := changeFiles(t, true)
	want := []string{"apply_test.go", "change_test.go", "composition_test.go"}
	if !slices.Equal(got, want) {
		t.Fatalf("change test files = %v, want %v", got, want)
	}
}

func TestChangeHasExactPublicSurface(t *testing.T) {
	tests := map[string][]string{
		"apply.go":       {"Apply", "ApplyAll"},
		"change.go":      {"Change", "Identity"},
		"composition.go": {"Compose", "Pipe"},
	}
	for filename, want := range tests {
		t.Run(filename, func(t *testing.T) {
			got := exportedNamesInSource(t, filepath.Join("..", "..", "change", filename))
			if !slices.Equal(got, want) {
				t.Fatalf("exports in %s = %v, want %v", filename, got, want)
			}
		})
	}
}

func changeFiles(t *testing.T, testFiles bool) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("..", "..", "change"))
	if err != nil {
		t.Fatalf("read change package: %v", err)
	}

	var names []string
	for _, entry := range entries {
		name := entry.Name()
		isTest := strings.HasSuffix(name, "_test.go")
		if !entry.IsDir() && strings.HasSuffix(name, ".go") && isTest == testFiles {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
