# Package `iterator`

**Import path:** `github.com/alvadorncorp/gopackages/functional/iterator`

`iterator` is a thin adapter and terminal layer over Go's `iter.Seq`; it does
not define an Iterator type. FromSlice delegates to slices.Values and Collect
delegates to slices.Collect. Map, Filter, FilterMap, FlatMap, Flatten, Take,
Drop, TakeWhile, DropWhile, Zip, Enumerate, Repeat, Iterate, and Unfold do no
work until iteration and materialize no intermediate collection.

Take(n<=0) does not start the source. Drop(n<=0) returns it unchanged. Any,
All, Find, and First stop upstream as soon as their answer is known. Empty Any
is false; empty All is true; absent Find/First/Last return zero,false.

Repeat and Iterate are infinite. Collect, Fold, Count, and Last over an
unbounded sequence never return unless a limiting adapter precedes them.

Zip uses iter.Pull, stops both pulls, and yields only complete Tuple2 values up
to the shorter input. Pull order is left then right; when right ends first, at
most one unmatched left value may already have been consumed. Source, callback,
downstream, and cleanup panics propagate; iterator does not recover them.

Production iterator imports only tuple, for Zip and Enumerate. It does not
import persistent, state, fold, Optional, reflect, unsafe, or external modules.
See [module overview](../overview.md) and [examples](../examples.md).
