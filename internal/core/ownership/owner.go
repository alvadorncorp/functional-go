package ownership

// Owner is a mutation-window token. Equality is pointer identity, and the
// token is only valid while Active. After Invalidate (Freeze), the same
// pointer is no longer an editable owner: nodes may still hold the stamp,
// but IsOwnedBy reports false so later windows must copy-on-write.
//
// The zero value is never obtained via New; nodes hold *Owner and treat
// nil as published / not exclusively owned (born published).
type Owner struct {
	active bool
}

// New returns a unique, active owner. Each call yields a distinct pointer.
func New() *Owner {
	return &Owner{active: true}
}

// Active reports whether o is a non-nil token that has not been invalidated.
func (o *Owner) Active() bool {
	return o != nil && o.active
}

// Invalidate permanently deactivates the token. It is idempotent and a
// no-op on a nil receiver. Freeze uses this so O(1) publish does not walk
// nodes: any remaining stamp of this pointer fails IsOwnedBy.
func (o *Owner) Invalidate() {
	if o == nil {
		return
	}
	o.active = false
}

// IsOwnedBy reports whether nodeOwner is the same live token as owner.
// It returns false when either pointer is nil, the pointers differ, or
// owner is inactive. (nil, nil) is not owned: published nodes are never
// editable in place.
func IsOwnedBy(nodeOwner, owner *Owner) bool {
	return nodeOwner != nil && owner != nil && nodeOwner == owner && owner.active
}
