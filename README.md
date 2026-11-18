# ![functional-go](docs/assets/logo.svg)

Functional programming building blocks for Go, with no third-party
dependencies. Requires **Go 1.27+**.

Import the subpackages you need: `go.alvadorn.org/functional` itself exports
nothing.

## Quick start

In a Go module, add the package:

```sh
go get go.alvadorn.org/functional/primitives/function
```

Save this as `main.go` and run `go run .`:

```go
package main

import (
	"fmt"
	"strings"

	"go.alvadorn.org/functional/primitives/function"
)

func main() {
	normalize := function.Pipe(strings.TrimSpace, strings.ToLower)
	fmt.Println(normalize("  GO  ")) // go
}
```

## Packages

| Area | What it provides |
|---|---|
| [`primitives`](primitives/) | Function composition, predicates, ordering, and folds |
| [`structural`](structural/) | Tuples, lazy values, iterators, and optics |
| [`values`](values/) | `Optional`, `Result`, `Either`, `NonEmpty`, and `Validated` |
| [`immutable`](immutable/) | Persistent lists, maps, sets, and sorted maps |
| [`state`](state/) | Pure state evolution and diffs |
| [`modeling`](modeling/) | Typed domain decisions and transitions |
| [`workflow`](workflow/) | Composition and execution of deferred effects; separate Go module |

Collections are shallowly immutable. `Optional[T]` has no built-in JSON
support: `encoding/json` encodes it as `{}`.

## Documentation

[Overview](docs/overview.md) · [More examples](docs/examples.md) ·
[Package guides and architecture decisions](docs/README.md)

## Development

From the repository root, run `make test-unit` and `make lint`.
Use `make test-published` to test without the workspace.

[GitHub Actions](.github/workflows/ci.yml) runs the Make targets `test-unit`,
`test-race`, `test-published`, `vet`, and `lint` in parallel for each module.
Run `make bench` separately for benchmarks.

Licensed under the [Apache License 2.0](LICENSE).
