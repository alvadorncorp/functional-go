package archtest

import (
	"slices"
	"testing"
)

var expectedPrimitiveAPI = map[string][]string{
	pkgFunction:  {"Compose", "Constant", "Flip", "Identity", "Pipe", "Tap"},
	pkgPredicate: {"All", "And", "Any", "EqualTo", "None", "Not", "NotEqualTo", "OneOf", "Or", "Predicate"},
	pkgOrder:     {"By", "Compare", "Equal", "Greater", "Less", "Max", "Min", "Natural", "Order", "Reverse", "Then"},
	pkgFold:      {"FoldLeft", "FoldRight", "Reduce", "ReduceNonEmpty", "ScanLeft"},
}

func TestPrimitivesPublicAPI(t *testing.T) {
	t.Parallel()
	for _, pkg := range primitivePackages {
		pkg := pkg
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			got := valuesExportedNames(t, pkg)
			want := slices.Clone(expectedPrimitiveAPI[pkg])
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("exported API for %s = %v, want %v", pkg, got, want)
			}
		})
	}
}
