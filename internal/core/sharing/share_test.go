package sharing_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
	"github.com/alvadorncorp/gopackages/functional/internal/core/sharing"
)

func TestPublished(t *testing.T) {
	t.Parallel()

	if !sharing.Published(nil) {
		t.Fatal("nil must be published")
	}
	if sharing.Published(ownership.New()) {
		t.Fatal("non-nil owner must not be published")
	}
}

func TestDetachClearsStamp(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	stamp := o
	sharing.Detach(&stamp)
	if stamp != nil {
		t.Fatal("Detach must set stamp to nil")
	}
	if !sharing.Published(stamp) {
		t.Fatal("detached stamp must be published")
	}
	if ownership.IsOwnedBy(stamp, o) {
		t.Fatal("detached node must not remain owned")
	}
}

func TestDetachNilPointerNoOp(t *testing.T) {
	t.Parallel()

	sharing.Detach(nil) // must not panic

	var stamp *ownership.Owner
	sharing.Detach(&stamp)
	if stamp != nil {
		t.Fatal("already-nil stamp must stay nil")
	}
}
