package archtest

import "testing"

func TestDependencyGateCanFail(t *testing.T) {
	if dependencyAllowed(&listedPackage{ImportPath: "example.com/unapproved"}) {
		t.Fatal("dependency gate accepted an external package")
	}
	for _, imported := range []string{
		"reflect",
		"unsafe",
		functionalModule,
		functionalModule + "/internal/core",
		functionalModule + "/values/result",
	} {
		if !directImportForbidden(modelingModule, imported) {
			t.Errorf("root direct-import gate accepted %s", imported)
		}
	}
	if directImportForbidden(internalDecisionPackage, functionalImmutable) {
		t.Fatalf("internal decision gate rejected %s", functionalImmutable)
	}
	for _, imported := range []string{functionalPredicate, functionalOptional} {
		if directImportForbidden(policyPackage, imported) {
			t.Errorf("policy direct-import gate rejected %s", imported)
		}
	}
	for _, imported := range []string{functionalImmutable, functionalModule + "/values/result"} {
		if !directImportForbidden(policyPackage, imported) {
			t.Errorf("policy direct-import gate accepted %s", imported)
		}
	}
	if directImportForbidden(changePackage, functionalFunction) {
		t.Fatalf("change direct-import gate rejected %s", functionalFunction)
	}
	for _, imported := range []string{
		functionalImmutable,
		functionalOptional,
		functionalPredicate,
		functionalModule + "/values/result",
	} {
		if !directImportForbidden(changePackage, imported) {
			t.Errorf("change direct-import gate accepted %s", imported)
		}
	}
	for _, pkg := range []string{reducerPackage, projectionPackage} {
		if directImportForbidden(pkg, functionalImmutable) {
			t.Errorf("%s direct-import gate rejected %s", pkg, functionalImmutable)
		}
		for _, imported := range []string{
			functionalOptional,
			functionalPredicate,
			functionalFunction,
			functionalModule + "/primitives/fold",
		} {
			if !directImportForbidden(pkg, imported) {
				t.Errorf("%s direct-import gate accepted %s", pkg, imported)
			}
		}
	}
	if _, forbidden := forbiddenExportedNames["Dispatch"]; !forbidden {
		t.Fatal("runtime API gate cannot reject Dispatch")
	}
}
