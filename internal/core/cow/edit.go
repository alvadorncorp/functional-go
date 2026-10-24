package cow

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

// EnsureEditable returns node when it is already exclusively owned by owner.
// Otherwise it returns the result of clone, which the caller must implement
// to allocate a structure-specific copy stamped with owner.
//
// Core never deep-copies and never inspects N. clone must not be nil when a
// copy is required; a nil clone in that case panics with
// "cow: clone required but nil".
func EnsureEditable[N any](
	nodeOwner *ownership.Owner,
	owner *ownership.Owner,
	node N,
	clone func() N,
) N {
	if ownership.IsOwnedBy(nodeOwner, owner) {
		return node
	}
	if clone == nil {
		panic("cow: clone required but nil")
	}
	return clone()
}
