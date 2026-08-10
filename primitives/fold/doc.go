// Package fold provides iterative left/right folds, typed reductions, and a
// preallocated left scan over slices. It does not depend on immutable
// collections; only Reduce and ReduceNonEmpty use the stable values Optional
// and NonEmpty contracts.
package fold
