package archtest

import "testing"

// TestNoForbiddenExportedNames proves no package exports a name that would
// promise an engine, a container, a registry, a resilience policy, or
// distributed atomicity (R7).
func TestNoForbiddenExportedNames(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, name := range productionNames(t, &production) {
			if _, forbidden := forbiddenExportedNames[name]; forbidden {
				t.Errorf("%s exports forbidden runtime API %s", production.ImportPath, name)
			}
		}
	}
}
