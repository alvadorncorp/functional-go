package ownership

// Owner is a mutation-window token. Equality is pointer identity only.
// The zero value is never obtained via New; nodes hold *Owner and treat
// nil as published / not exclusively owned.
//
// The unexported byte keeps Owner non-zero-size so each New() allocation
// has a distinct address (Go collapses all zero-sized allocations).
type Owner struct{ _ byte }

// New returns a unique owner. Each call yields a distinct pointer identity.
func New() *Owner {
	return new(Owner)
}

// IsOwnedBy reports whether nodeOwner is the same token as owner.
// It returns false when either pointer is nil or the pointers differ.
// (nil, nil) is not owned: published nodes are never editable in place.
func IsOwnedBy(nodeOwner, owner *Owner) bool {
	return nodeOwner != nil && owner != nil && nodeOwner == owner
}
