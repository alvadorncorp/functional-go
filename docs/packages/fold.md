# Package `primitives/fold`

**Import path:** `go.alvadorn.org/functional/primitives/fold`

`FoldLeft` and `FoldRight` reduce slices with an explicit initial state.
FoldRight uses a reverse loop rather than recursion. `Reduce` uses the first
slice element as state and returns `values/optional.Optional[T]` so empty input
is represented in the type. `ReduceNonEmpty` consumes
`values/nonempty.NonEmpty[T]` and therefore returns `T` directly.

`ScanLeft` includes the initial state, so its result length is always
`len(values)+1`; it allocates exactly the output slice. FoldLeft, FoldRight,
Reduce, and ReduceNonEmpty allocate no intermediate collection.

These folds take **slices**. To reduce an `iter.Seq`, use
[`iterator.Fold`](iterator.md#fold-over-iterseq-vs-slice-folds): it consumes
the sequence once, left to right, with no intermediate slice. Choose by what
you already hold: a slice goes to `fold.*`, a sequence goes to
`iterator.Fold`; avoid converting one into the other just to fold.

The package imports exactly `values/optional` and `values/nonempty`. It does not
import immutable collections or the other primitive packages. See
[functional primitives and fold boundary](../adr/functional/functional-primitives-and-fold-boundary.md).
