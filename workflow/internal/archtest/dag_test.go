package archtest

import (
	"strings"
	"testing"
)

// allowedInternalEdges is the intra-module production dependency DAG. action is
// the sink; every other package may point at it and nothing may point back.
// A package missing from this map is an unreviewed addition and fails the gate.
func allowedInternalEdges() map[string]map[string]struct{} {
	return map[string]map[string]struct{}{
		workflowModule:     {actionPackage: {}},
		actionPackage:      {},
		transactionPackage: {actionPackage: {}},
	}
}

// TestProductionDependencyDAG proves the module's internal edges are exactly
// the reviewed ones, so action -> workflow can never appear (R3).
func TestProductionDependencyDAG(t *testing.T) {
	edges := allowedInternalEdges()

	for _, production := range productionPackages(t) {
		allowed, known := edges[production.ImportPath]
		if !known {
			t.Errorf("%s is not declared in allowedInternalEdges", production.ImportPath)
			continue
		}
		for _, imported := range production.Imports {
			if !strings.HasPrefix(imported, workflowModule) {
				continue
			}
			if _, permitted := allowed[imported]; !permitted {
				t.Errorf("forbidden internal edge %s -> %s", production.ImportPath, imported)
			}
		}
	}
}
