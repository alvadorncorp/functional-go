package sharing

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

// Detach clears *nodeOwner to nil, marking the node published/shared.
// After Detach, ownership.IsOwnedBy is always false for that stamp; any
// later mutation under any owner requires a structure-defined copy.
// Detach is a no-op when nodeOwner is nil or *nodeOwner is already nil.
func Detach(nodeOwner **ownership.Owner) {
	if nodeOwner == nil {
		return
	}
	*nodeOwner = nil
}

// Published reports whether nodeOwner means shared immutable (nil stamp).
func Published(nodeOwner *ownership.Owner) bool {
	return nodeOwner == nil
}
