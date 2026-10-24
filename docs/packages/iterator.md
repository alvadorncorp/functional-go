# Package `structural/iterator`

**Import path:** `github.com/alvadorncorp/gopackages/functional/structural/iterator`

`iterator` is a thin adapter and terminal layer over Go's `iter.Seq`; it does
not define an Iterator type. Map, Filter, FilterMap, FlatMap, Flatten, Take,
Drop, TakeWhile, DropWhile, Zip, Enumerate, Repeat, Iterate, and Unfold do no
work until iteration and materialize no intermediate collection.

Take(n<=0) does not start the source. Drop(n<=0) returns it unchanged. Any,
All, Find, and First stop upstream as soon as their answer is known. Empty Any
is false; empty All is true; absent Find/First/Last return zero,false.

Repeat and Iterate are infinite. `slices.Collect`, Fold, Count, and Last over an
unbounded sequence never return unless a limiting adapter precedes them.

Zip uses iter.Pull, stops both pulls, and yields only complete Tuple2 values up
to the shorter input. Pull order is left then right; when right ends first, at
most one unmatched left value may already have been consumed. Source, callback,
downstream, and cleanup panics propagate; iterator does not recover them.

Production iterator imports only tuple, for Zip and Enumerate. It does not
import immutable, state, fold, Optional, reflect, unsafe, or external modules.

## Deprecated synonyms (kept in v1)

`FromSlice` and `FromValues` only forward to `slices.Values`; `Collect` only
forwards to `slices.Collect`. All three are marked `Deprecated:` and keep
working with unchanged behavior. Prefer the standard library:

| Deprecated | Use |
|---|---|
| `iterator.FromSlice(values)` | `slices.Values(values)` |
| `iterator.FromValues(a, b, c)` | `slices.Values([]T{a, b, c})` |
| `iterator.Collect(seq)` | `slices.Collect(seq)` |

The slice sources are **live views**: `slices.Values` does not copy, so a
consumer that iterates after the caller mutates the slice sees the new
contents. They are not immutable snapshots. See
[ADR-014](../adr/functional/adr-014-v1-surface-hygiene.md).

## Fold over `iter.Seq` vs. slice folds

`iterator.Fold(seq, initial, fn)` reduces any `iter.Seq` left to right,
consumes it once, and is the fold to use when the source is already a
sequence. `primitives/fold` (`FoldLeft`, `FoldRight`, `ScanLeft`, `Reduce`)
reduces **slices**: use it when you hold a slice, need right-to-left order,
scan states, or optional/non-empty results. Do not turn a slice into a
sequence only to call `iterator.Fold`, and do not collect a sequence into a
slice only to call `fold.FoldLeft`. `iterator` does not import `fold`.

See [module overview](../overview.md) and [examples](../examples.md).
