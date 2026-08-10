package archtest

import (
	"slices"
	"strings"
	"testing"
)

const internalDecisionPackage = modelingModule + "/internal/decision"

func TestDecisionDependencyDirection(t *testing.T) {
	packages := make(map[string]listedPackage)
	for _, pkg := range productionPackages(t) {
		packages[pkg.ImportPath] = pkg
	}

	root, rootFound := packages[modelingModule]
	if !rootFound {
		t.Fatalf("production packages missing %s", modelingModule)
	}
	internal, internalFound := packages[internalDecisionPackage]
	if !internalFound {
		t.Fatalf("production packages missing %s", internalDecisionPackage)
	}
	if !slices.Contains(root.Imports, internalDecisionPackage) {
		t.Errorf("%s must import %s", modelingModule, internalDecisionPackage)
	}
	if slices.Contains(root.Imports, functionalImmutable) {
		t.Errorf("%s must delegate immutable storage through %s", modelingModule, internalDecisionPackage)
	}
	if !slices.Contains(internal.Imports, functionalImmutable) {
		t.Errorf("%s must import %s", internalDecisionPackage, functionalImmutable)
	}
	for _, imported := range internal.Imports {
		if imported == modelingModule || strings.HasPrefix(imported, modelingModule+"/") {
			t.Errorf("%s must not import modeling package %s", internalDecisionPackage, imported)
		}
	}
}
