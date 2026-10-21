package wbtree

import (
	"cmp"
	"maps"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// maxFuzzOps bounds how many instructions one fuzz input decodes into.
const maxFuzzOps = 200

// fuzzOpWidth is how many bytes each instruction consumes: one opcode byte
// plus key and value bytes (value unused by deletes).
const fuzzOpWidth = 3

// FuzzOps runs a small program of pure and transient Set/Delete operations
// against a Root and a map[int]int model, checking the tree invariants and
// the ascending contents after every step. Pure steps first freeze any open
// owner, as a builder's Build does, and every published version must still
// match its model snapshot at the end.
func FuzzOps(f *testing.F) {
	f.Add([]byte{0, 1, 10, 0, 2, 20, 1, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{2, 5, 1, 2, 6, 2, 3, 5, 0, 0, 7, 3, 3, 6, 0})
	f.Add([]byte{2, 1, 1, 2, 2, 2, 2, 3, 3, 2, 4, 4, 1, 2, 0, 3, 1, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			root  Root[int, int]
			model map[int]int
		}

		model := map[int]int{}
		current := New[int, int](cmp.Compare[int])
		var owner *ownership.Owner
		var published []version

		ops := 0
		for cur := 0; cur+fuzzOpWidth <= len(data) && ops < maxFuzzOps; cur += fuzzOpWidth {
			ops++
			op, k, v := data[cur]%4, int(data[cur+1]), int(data[cur+2])

			if op < 2 && owner != nil { // freeze before a pure step
				owner.Invalidate()
				owner = nil
				published = append(published, version{current, maps.Clone(model)})
			}
			if op >= 2 && owner == nil {
				owner = ownership.New()
			}
			switch op {
			case 0:
				current = current.Set(k, v)
				model[k] = v
			case 1:
				current = current.Delete(k)
				delete(model, k)
			case 2:
				current = current.TransientSet(owner, k, v)
				model[k] = v
			case 3:
				current = current.TransientDelete(owner, k)
				delete(model, k)
			}
			if owner == nil {
				published = append(published, version{current, maps.Clone(model)})
			}
			checkInvariants(t, current)
			assertMatchesModel(t, current, model)
		}
		owner.Invalidate()

		for i, v := range published {
			checkInvariants(t, v.root)
			if got := maps.Collect(v.root.All()); !maps.Equal(got, v.model) {
				t.Fatalf("published version[%d] = %v, want %v; a later operation reached back into it",
					i, got, v.model)
			}
		}
	})
}

// assertMatchesModel fails t unless r yields exactly model's entries in
// ascending key order.
func assertMatchesModel(t *testing.T, r Root[int, int], model map[int]int) {
	t.Helper()
	if r.Len() != len(model) {
		t.Fatalf("Len = %d, want %d", r.Len(), len(model))
	}
	keys := slices.Sorted(maps.Keys(model))
	i := 0
	for k, v := range r.All() {
		if i >= len(keys) || k != keys[i] || v != model[k] {
			t.Fatalf("All()[%d] = (%d, %d), want keys %v of model %v", i, k, v, keys, model)
		}
		i++
	}
	if i != len(keys) {
		t.Fatalf("All yielded %d entries, want %d", i, len(keys))
	}
}
