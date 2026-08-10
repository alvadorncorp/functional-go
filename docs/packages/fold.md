# Package `primitives/fold`

**Import path:** `github.com/alvadorncorp/gopackages/functional/primitives/fold`

`FoldLeft` and `FoldRight` reduce slices with an explicit initial state.
FoldRight uses a reverse loop rather than recursion. `Reduce` uses the first
slice element as state and returns `values/optional.Optional[T]` so empty input
is represented in the type. `ReduceNonEmpty` consumes
`values/nonempty.NonEmpty[T]` and therefore returns `T` directly.

`ScanLeft` includes the initial state, so its result length is always
`len(values)+1`; it allocates exactly the output slice. FoldLeft, FoldRight,
Reduce, and ReduceNonEmpty allocate no intermediate collection.

The package imports exactly `values/optional` and `values/nonempty`. It does not
import immutable collections or the other primitive packages. See
[ADR-008](../adr/functional/adr-008-functional-primitives-and-fold-boundary.md).
