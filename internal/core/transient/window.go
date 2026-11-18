package transient

import "go.alvadorn.org/functional/internal/core/ownership"

// panicUsedAfterFreeze is the stable message for a nil window, a zero-value
// window, or any use after Freeze.
const panicUsedAfterFreeze = "transient: window used after Freeze"

// Window is a single-threaded mutation scope. It is not safe for concurrent use.
// An open window holds a non-nil owner; a closed, nil, or zero-value window has
// no owner and panics on Owner or Freeze.
type Window struct {
	owner *ownership.Owner
}

// Open creates a Window with a fresh ownership.Owner.
func Open() *Window {
	return &Window{owner: ownership.New()}
}

// Owner returns the live token for this window.
// It panics with "transient: window used after Freeze" if w is nil, zero-value,
// or already frozen.
func (w *Window) Owner() *ownership.Owner {
	if w == nil || w.owner == nil {
		panic(panicUsedAfterFreeze)
	}
	return w.owner
}

// Freeze permanently closes the window in O(1): it Invalidates the owner
// token (so any node still stamped with that *Owner fails IsOwnedBy) and
// clears the window's owner pointer. It does not walk nodes and does not
// require Detach. A second Freeze, a nil receiver, or a zero-value window
// panics with "transient: window used after Freeze".
func (w *Window) Freeze() {
	if w == nil || w.owner == nil {
		panic(panicUsedAfterFreeze)
	}
	w.owner.Invalidate()
	w.owner = nil
}
