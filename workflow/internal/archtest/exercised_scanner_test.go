package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func parseConsumerFixture(t *testing.T, source string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "consumer.go", source, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return file
}

func TestExactImportedSelectors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []surfaceSymbol
		badDot bool
	}{
		{
			"default alias", `import "` + actionPackage + `"; var _ = action.Map[int,int]`,
			[]surfaceSymbol{{actionPackage, "Map"}},
			false,
		},
		{
			"renamed alias and generic", `import a "` + actionPackage + `"; var _ = a.Zip[int,string]`,
			[]surfaceSymbol{{actionPackage, "Zip"}},
			false,
		},
		{
			"exact prefix", `import "` + workflowModule + `"; var _ = workflow.MapFallible[int,int,int]`,
			[]surfaceSymbol{{workflowModule, "MapFallible"}},
			false,
		},
		{"comment and string", `// action.Map workflow.Map
var _ = "action.Attempt"`, nil, false},
		{"foreign same qualifier", `import action "example.org/action"; var _ = action.Map`, nil, false},
		{"foreign path suffix", `import "` + actionPackage + `/other"; var _ = action.Map`, nil, false},
		{"blank", `import _ "` + actionPackage + `"; var _ = action.Map[int,int]`, nil, false},
		{"tracked dot rejected", `import . "` + actionPackage + `"; var _ = Map[int,int]`, nil, true},
		{"foreign dot ignored", `import . "example.org/action"; var _ = Map`, nil, false},
		{"all package defaults", `import (
"` + workflowModule + `"
"` + actionPackage + `"
"` + planPackage + `"
"` + resiliencePackage + `"
"` + transactionPackage + `"
)
var _ workflow.Workflow[int,int]
var _ action.Action[int]
var _ = plan.KindStep
var _ resilience.RetryPolicy
var _ transaction.Boundary`, []surfaceSymbol{
			{workflowModule, "Workflow"},
			{actionPackage, "Action"},
			{planPackage, "KindStep"},
			{resiliencePackage, "RetryPolicy"},
			{transactionPackage, "Boundary"},
		}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumed := map[surfaceSymbol]bool{}
			err := collectSurfaceSelectors(parseConsumerFixture(t, "package consumer\n"+test.source), consumed)
			if (err != nil) != test.badDot {
				t.Fatalf("selector scan error = %v, want dot rejection %v", err, test.badDot)
			}
			assertConsumedSymbols(t, consumed, test.want)
		})
	}
}

func assertConsumedSymbols(t *testing.T, got map[surfaceSymbol]bool, want []surfaceSymbol) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("consumed = %v, want %v", got, want)
	}
	for _, symbol := range want {
		if !got[symbol] {
			t.Errorf("missing %v in %v", symbol, got)
		}
	}
}

func TestImportAliasLexicalBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		mapUsed bool
	}{
		{"parameter shadows body", `func f(a struct{Map int}) { _ = a.Map }`, false},
		{"parameter not signature", `func f(a struct{Map int}, _ a.Action[int]) { _ = a.Map }`, false},
		{"signature uses import", `func f(a struct{Map int}, _ a.Action[int]) a.Action[int] {
_ = a.Map; return nil
}`, false},
		{"signature selector before named result scope", `func f() (a a.Action[int]) { return }`, false},
		{"named result shadows body", `func f() (a struct{Map int}) { _ = a.Map; return }`, false},
		{"receiver shadows body", `type local struct{Map int}; func (a local) f() { _ = a.Map }`, false},
		{"receiver not signature", `type local struct{};
func (a local) f() a.Action[int] { return nil }`, false},
		{"func literal parameter", `func f() { _ = func(a struct{Map int}) { _ = a.Map } }`, false},
		{"closure inherits outer", `func f(a struct{Map int}) { _ = func() { _ = a.Map } }`, false},
		{"closure binding does not escape", `func f() {
_ = func(a struct{Map int}) { _ = a.Map }; _ = a.Map[int,int]
}`, true},
		{"short declaration shadows after", `func f() { a := struct{Map int}{}; _ = a.Map }`, false},
		{"before declaration", `func f() { _ = a.Map[int,int]; a := struct{Map int}{}; _ = a.Map }`, true},
		{"own short initializer", `func f() { a := struct{Map any}{a.Map[int,int]}; _ = a.Map }`, true},
		{"var shadows after", `func f() { var a struct{Map int}; _ = a.Map }`, false},
		{"own var initializer", `func f() { var a = struct{Map any}{a.Map[int,int]}; _ = a.Map }`, true},
		{"grouped spec starts immediately", `func f() {
var (a = struct{Map int}{}; b = a.Map); _ = b
}`, false},
		{"grouped own initializer", `func f() {
var (a = struct{Map any}{a.Map[int,int]}; b = a.Map); _ = b
}`, true},
		{"const shadows later spec", `func f() { const (a = 0; b = a.Map); _ = b }`, false},
		{"assignment declares nothing", `func f() { var x any; x = a.Map[int,int]; _ = x }`, true},
		{"restored after block", `func f() {
{ a := struct{Map int}{}; _ = a.Map }; _ = a.Map[int,int]
}`, true},
		{"if init covers else", `func f() {
if a := (struct{Map int}{}); a.Map == 0 { _ = a.Map } else { _ = a.Map }
}`, false},
		{"if own init", `func f() { if a := (struct{Map any}{a.Map[int,int]}); a.Map != nil {} }`, true},
		{"if ends", `func f() { if a := (struct{Map int}{}); a.Map == 0 {}; _ = a.Map[int,int] }`, true},
		{"for init covers condition post body", `func f() {
for a := (struct{Map int}{}); a.Map < 1; a.Map++ { _ = a.Map }
}`, false},
		{"for ends", `func f() {
for a := (struct{Map int}{}); a.Map < 1; a.Map++ {}; _ = a.Map[int,int]
}`, true},
		{"switch init", `func f() { switch a := (struct{Map int}{}); a.Map { case 0: _ = a.Map } }`, false},
		{"switch clause isolated", `func f(x int) {
switch x { case 0: a := struct{Map int}{}; _ = a.Map; case 1: _ = a.Map[int,int] }
}`, true},
		{"range body", `func f(xs []struct{Map int}) { for _, a := range xs { _ = a.Map } }`, false},
		{"range RHS keeps import", `func f() {
for _, a := range []struct{Map any}{{a.Map[int,int]}} { _ = a.Map }
}`, true},
		{"range ends", `func f(xs []struct{Map int}) {
for _, a := range xs { _ = a.Map }; _ = a.Map[int,int]
}`, true},
		{"range assignment not declaration", `func f(xs []int) {
var x int; for _, x = range xs { _ = a.Map[int,int] }
}`, true},
		{"select clause receive", `func f(ch chan struct{Map int}) {
select { case a := <-ch: _ = a.Map; default: }
}`, false},
		{"select sibling restored", `func f(ch chan struct{Map int}) {
select { case a := <-ch: _ = a.Map; default: _ = a.Map[int,int] }
}`, true},
		{"receive RHS before scope", `func f() {
select { case a := <-make(chan struct{Map any}, len([]any{a.Map[int,int]})): _ = a.Map }
}`, true},
		{"type switch bodies", `func f(x any) {
switch a := x.(type) { case struct{Map int}: _ = a.Map; default: _ = a }
}`, false},
		{"type switch case types before binding", `func f(x any) {
switch a := x.(type) { case a.Action[int]: _ = a; default: _ = a }
}`, false},
		{"type switch guard RHS", `func f() {
switch a := any(a.Map[int,int]).(type) { default: _ = a }
}`, true},
		{"type switch ends", `func f(x any) {
switch a := x.(type) { default: _ = a }; _ = a.Map[int,int]
}`, true},
		{"function type parameter", `func f[a interface{Map()}]() { _ = a.Map }`, false},
		{"type parameter constraint", `type Box[a interface{ ~[]a.Map }] struct{}`, false},
		{"generic receiver parameter", `type Box[T any] struct{}; func (b Box[a]) f() { _ = a.Map }`, false},
		{"generic pointer receiver parameters", `type Box[T,U any] struct{};
func (b *Box[a,T]) f() { _ = a.Map }`, false},
		{"local type scope", `func f() { type a struct{Map int}; _ = a{}.Map }`, false},
		{"local type begins at name", `func f() { type a struct{X a.Map}; _ = a{} }`, false},
		{"local type ends", `func f() {
{ type a struct{Map int}; _ = a{}.Map }; _ = a.Map[int,int]
}`, true},
		{"field names not bindings", `type local struct{a int}; var _ = a.Map[int,int]`, true},
		{"function type parameter names not bindings", `type local func(a int);
var _ = a.Map[int,int]`, true},
		{"label not binding", `func f() { a: for { _ = a.Map[int,int]; break a } }`, true},
		{"literal key not binding", `func f() { _ = struct{a any}{a: a.Map[int,int]} }`, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "package consumer\nimport a \"" + actionPackage + "\"\nvar _ a.Action[int]\n" + test.body
			consumed := map[surfaceSymbol]bool{}
			if err := collectSurfaceSelectors(parseConsumerFixture(t, source), consumed); err != nil {
				t.Fatal(err)
			}
			want := []surfaceSymbol{{actionPackage, "Action"}}
			if test.mapUsed {
				want = append(want, surfaceSymbol{actionPackage, "Map"})
			}
			assertConsumedSymbols(t, consumed, want)
		})
	}
}

func TestConsumerFileBoundary(t *testing.T) {
	for _, test := range []struct {
		path     string
		examples bool
		want     bool
	}{
		{"compose.go", false, true},
		{"action/construct.go", false, true},
		{"example_test.go", false, true},
		{"action/example_test.go", false, true},
		{"action/action_test.go", false, false},
		{"internal/archtest/exercised_test.go", false, false},
		{"internal/archtest/example_test.go", false, false},
		{"docs/example.md", false, false},
		{"checkout/main.go", true, true},
		{"checkout/main_test.go", true, true},
	} {
		t.Run(test.path, func(t *testing.T) {
			if got := consumerFile(filepath.FromSlash(test.path), test.examples); got != test.want {
				t.Errorf("consumerFile(%q, %v) = %v, want %v", test.path, test.examples, got, test.want)
			}
		})
	}
}

func TestConsumerWalkUsesRelativeModuleBoundary(t *testing.T) {
	root := filepath.Join(t.TempDir(), "examples", "checkout", "workflow")
	if err := os.MkdirAll(root, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"ordinary_test.go": "package consumer\nimport a \"" + actionPackage + "\"\nvar _ = a.Map[int,int]",
		"example_test.go":  "package consumer\nimport a \"" + actionPackage + "\"\nvar _ = a.Zip[int,int]",
		"sibling.go":       "package consumer\nimport a \"" + actionPackage + "\"\nvar _ a.Action[int]",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	consumed := map[surfaceSymbol]bool{}
	scanConsumerTree(t, root, false, consumed)
	assertConsumedSymbols(t, consumed, []surfaceSymbol{{actionPackage, "Zip"}, {actionPackage, "Action"}})
	consumed = map[surfaceSymbol]bool{}
	scanConsumerTree(t, root, true, consumed)
	assertConsumedSymbols(t, consumed, []surfaceSymbol{
		{actionPackage, "Map"}, {actionPackage, "Zip"}, {actionPackage, "Action"},
	})
}

func TestSurfaceResultAndFieldReachability(t *testing.T) {
	file := parseConsumerFixture(t, `package action
import foreign "time"
type Action[T any] func(int) (T,error)
type Pair[A,B any] struct{ First A; Second B }
type AttemptResult[T any] struct{ Value T; Err error }
type Hidden struct{}
type FieldNameOnly struct{}
type ParameterOnly struct{}
type BodyOnly struct{}
type Duration struct{}
type TypeParameter struct{}
type MethodOnly struct{}
type Backoff func(int) int
type privateLink struct{ Next *Backoff }
type Policy struct{ FieldNameOnly int; Public []privateLink; private Hidden }
type Cycle struct{ Next *Cycle }
type Other struct{}
func Zip[A,B any](a Action[A], b Action[B]) Action[Pair[A,B]] { return nil }
func Attempt[T any](a Action[T]) Action[AttemptResult[T]] { return nil }
func Unconsumed() Hidden { return Hidden{} }
func Map(a ParameterOnly) Action[int] { _ = BodyOnly{}; return nil }
func Foreign() foreign.Duration { panic(0) }
func Generic[TypeParameter any]() TypeParameter { panic(0) }
func ReturnsPrivate() privateLink { panic(0) }
func (p Policy) Method() MethodOnly { panic(0) }
func OtherFunction() Other { panic(0) }
`)
	graph := map[surfaceSymbol]surfaceTypeEdges{}
	addSurfaceTypeEdges(graph, actionPackage, file)
	tests := []struct {
		name  string
		roots []string
		want  []string
	}{
		{"Zip nested pair", []string{"Zip"}, []string{"Zip", "Action", "Pair"}},
		{"Attempt nested result", []string{"Attempt"}, []string{"Attempt", "Action", "AttemptResult"}},
		{"unrelated Map", []string{"Map"}, []string{"Map", "Action"}},
		{"policy public field graph", []string{"Policy"}, []string{"Policy", "privateLink", "Backoff"}},
		{"foreign selector parts", []string{"Foreign"}, []string{"Foreign"}},
		{"bound type parameter", []string{"Generic"}, []string{"Generic"}},
		{"private result type", []string{"ReturnsPrivate"}, []string{"ReturnsPrivate", "privateLink", "Backoff"}},
		{"cyclic type", []string{"Cycle"}, []string{"Cycle"}},
		{"no consumed functions", nil, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumed := map[surfaceSymbol]bool{}
			for _, root := range test.roots {
				consumed[surfaceSymbol{actionPackage, root}] = true
			}
			reachSurfaceTypes(consumed, graph)
			var want []surfaceSymbol
			for _, name := range test.want {
				want = append(want, surfaceSymbol{actionPackage, name})
			}
			assertConsumedSymbols(t, consumed, want)
		})
	}
}

func TestCarrierGraphFollowsOnlyTypePositions(t *testing.T) {
	file := parseConsumerFixture(t, `package example
type Item struct{}
type Key struct{}
type Ignored struct{}
type Alias = Container
type Container struct {
 Exported map[Key]chan *Item
 Callback func(Ignored) []Item
 Embedded
 private Ignored
}
type Embedded struct{}
type Interface interface { Embedded; Method() Ignored }
type NotATypeParameter[Item any] struct { Public Item }
type Array [arraySize]Item
const arraySize = 1
func Result() Alias { panic(0) }
func Signature() (named Item) { panic(0) }
func arraySizeFunction() Ignored { panic(0) }
`)
	graph := map[surfaceSymbol]surfaceTypeEdges{}
	addSurfaceTypeEdges(graph, actionPackage, file)
	for _, test := range []struct {
		root string
		want []string
	}{
		{"Result", []string{"Result", "Alias", "Container", "Key", "Item", "Embedded"}},
		{"Signature", []string{"Signature", "Item"}},
		{"Interface", []string{"Interface", "Embedded"}},
		{"NotATypeParameter", []string{"NotATypeParameter"}},
		{"Array", []string{"Array", "Item"}},
	} {
		t.Run(test.root, func(t *testing.T) {
			consumed := map[surfaceSymbol]bool{{actionPackage, test.root}: true}
			reachSurfaceTypes(consumed, graph)
			var got []string
			for symbol := range consumed {
				got = append(got, symbol.name)
			}
			slices.Sort(got)
			slices.Sort(test.want)
			if !slices.Equal(got, test.want) {
				t.Fatalf("reachable = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSignatureAndCaseSelectorsKeepImportBinding(t *testing.T) {
	for _, body := range []string{
		`func f(a struct{Map int}, _ a.Pair[int,int]) { _ = a.Map }`,
		`func f() (a a.Pair[int,int]) { return }`,
		`type local struct{}; func (a local) f() a.Pair[int,int] { panic(0) }`,
		`func f() { _ = func(a struct{Map int}) a.Pair[int,int] { _ = a.Map; panic(0) } }`,
		`func f(x any) { switch a := x.(type) { case a.Pair[int,int]: _ = a; default: _ = a } }`,
	} {
		t.Run(body, func(t *testing.T) {
			source := "package consumer\nimport a \"" + actionPackage + "\"\nvar _ a.Action[int]\n" + body
			consumed := map[surfaceSymbol]bool{}
			if err := collectSurfaceSelectors(parseConsumerFixture(t, source), consumed); err != nil {
				t.Fatal(err)
			}
			assertConsumedSymbols(t, consumed, []surfaceSymbol{
				{actionPackage, "Action"}, {actionPackage, "Pair"},
			})
		})
	}
}

func TestResultCarrierGraphAcrossFiles(t *testing.T) {
	graph := map[surfaceSymbol]surfaceTypeEdges{}
	for _, source := range []string{
		"package example; func Consumed() privateLink { panic(0) }; func Unconsumed() Unused { panic(0) }",
		"package example; type privateLink struct{ Next *Carrier }; type Carrier struct{}; type Unused struct{}",
	} {
		addSurfaceTypeEdges(graph, actionPackage, parseConsumerFixture(t, source))
	}
	consumed := map[surfaceSymbol]bool{{actionPackage, "Consumed"}: true}
	reachSurfaceTypes(consumed, graph)
	assertConsumedSymbols(t, consumed, []surfaceSymbol{
		{actionPackage, "Consumed"}, {actionPackage, "privateLink"}, {actionPackage, "Carrier"},
	})
}
