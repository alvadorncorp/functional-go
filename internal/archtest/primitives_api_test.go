package archtest

import (
	"reflect"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/tuple"
)

var expectedPrimitiveAPI = map[string][]string{
	pkgFunction:  {"Compose", "Constant", "Flip", "Identity", "Pipe", "Tap"},
	pkgPredicate: {"All", "And", "Any", "EqualTo", "None", "Not", "NotEqualTo", "OneOf", "Or", "Predicate"},
	pkgOrder:     {"By", "Compare", "Equal", "Greater", "Less", "Max", "Min", "Natural", "Order", "Reverse", "Then"},
	pkgFold:      {"FoldLeft", "FoldRight", "Reduce", "ReduceNonEmpty", "ScanLeft"},
	pkgTuple: {
		"FromValues2", "MapFirst", "MapFirst3", "MapSecond", "MapSecond3",
		"MapThird3", "Of2", "Of3", "Swap", "Tuple2", "Tuple3",
	},
	pkgLazy: {"FlatMap", "Get", "Lazy", "Map", "New", "Of"},
	pkgIterator: {
		"All", "Any", "Collect", "Count", "Drop", "DropWhile", "Empty",
		"Enumerate", "Filter", "FilterMap", "Find", "First", "FlatMap",
		"Flatten", "Fold", "FromSlice", "FromValues", "Iterate", "Last",
		"Map", "Once", "Repeat", "Take", "TakeWhile", "Unfold", "Zip",
	},
	pkgOptics: {"Compose", "Get", "Lens", "Modify", "NewLens", "Set"},
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

func TestTupleFieldShape(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		typeOf reflect.Type
		fields []string
	}{
		{reflect.TypeOf(tuple.Tuple2[int, string]{}), []string{"First", "Second"}},
		{reflect.TypeOf(tuple.Tuple3[int, string, bool]{}), []string{"First", "Second", "Third"}},
	} {
		if test.typeOf.NumField() != len(test.fields) {
			t.Fatalf("%s fields = %d, want %d", test.typeOf, test.typeOf.NumField(), len(test.fields))
		}
		for index, want := range test.fields {
			field := test.typeOf.Field(index)
			if field.Name != want || field.PkgPath != "" {
				t.Fatalf(
					"%s field %d = %s exported=%v, want %s exported",
					test.typeOf, index, field.Name, field.PkgPath == "", want,
				)
			}
		}
	}
}
