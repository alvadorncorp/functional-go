package hamt

import (
	"maps"
	"testing"

	"go.alvadorn.org/functional/internal/core/ownership"
)

func FuzzHAMTCollisions(f *testing.F) {
	f.Add([]byte{})
	// Pure collision growth, replacement, missing delete, >2 delete, and 2->leaf.
	f.Add([]byte{
		0, 0, 1, 0, 4, 2, 0, 8, 3, 0, 4, 9,
		1, 12, 0, 1, 8, 0, 1, 0, 0, 1, 4, 0,
	})
	// The same transitions inside one active-owner batch, then publication.
	f.Add([]byte{
		2, 0, 1, 2, 4, 2, 2, 8, 3, 2, 4, 9,
		3, 12, 0, 3, 8, 0, 3, 0, 0, 3, 4, 0, 4, 0, 0,
	})
	// Clone a published collision, then branch under fresh owners; a published
	// owner is invalidated and never reused.
	f.Add([]byte{
		0, 0, 1, 0, 4, 2, 0, 8, 3, 2, 4, 9, 2, 12, 4,
		3, 8, 0, 4, 0, 0, 5, 2, 0, 2, 0, 5, 2, 8, 6,
		4, 0, 0, 6, 4, 0, 0, 4, 7, 5, 2, 0, 3, 0, 0,
	})
	// All four hashes, collision-to-bitmap merging, and colliding bitmap children.
	f.Add([]byte{
		2, 0, 1, 2, 4, 2, 2, 1, 3, 2, 5, 4,
		2, 2, 5, 2, 6, 6, 2, 3, 7, 2, 7, 8,
		4, 0, 0, 0, 8, 9, 1, 5, 0, 2, 11, 10, 3, 7, 0,
	})
	// Divergent-hash and same-hash missing deletes on a collision root.
	f.Add([]byte{
		0, 0, 0, 0, 4, 1, 1, 1, 0, 1, 8, 0,
		3, 1, 0, 3, 8, 0, 2, 1, 2, 4, 0, 0,
	})
	// Pure merging from a collision root and a fresh-owner >2 collision delete.
	f.Add([]byte{
		0, 0, 1, 0, 4, 2, 0, 8, 3, 0, 1, 4,
		5, 3, 0, 3, 4, 0, 3, 0, 0, 2, 12, 5, 4, 0, 0,
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Bound work to 96 three-byte operations over 32 keys. Every published
		// snapshot is retained and checked after every operation, without sampling.
		data = data[:min(len(data), 96*3)]
		type version struct {
			node  node[int, int]
			model map[int]int
		}
		current := version{model: make(map[int]int)}
		history := []version{{model: make(map[int]int)}}
		var owner *ownership.Owner
		defer func() {
			if owner != nil {
				owner.Invalidate()
			}
		}()
		publish := func() {
			if owner != nil {
				owner.Invalidate()
				assertNotOwnedByInactive(t, current.node, owner)
				owner = nil
			}
			history = append(history, version{node: current.node, model: maps.Clone(current.model)})
		}
		check := func(step int) {
			t.Helper()
			assertCollisionModel(t, step, -1, current.node, current.model)
			for i, snapshot := range history {
				assertCollisionModel(t, step, i, snapshot.node, snapshot.model)
			}
		}
		check(-1)
		for i := 0; i+2 < len(data); i += 3 {
			op, key, value := data[i]%8, int(data[i+1]%32), int(data[i+2])
			switch op {
			case 0, 1, 2, 3:
				if op < 2 && owner != nil {
					publish()
				}
				if op >= 2 && owner == nil {
					owner = ownership.New()
				}
				hash := uint64(key % 4)
				_, existed := current.model[key]
				if op%2 == 0 {
					var added bool
					switch {
					case current.node == nil:
						current.node = &leaf[int, int]{owner: owner, hash: hash, key: key, val: value}
						added = true
					case owner == nil:
						current.node, added = current.node.set(hash, 0, key, value)
					default:
						current.node, added = current.node.tset(owner, hash, 0, key, value)
					}
					if added == existed {
						t.Fatalf("step %d: set key %d added=%v, previously present=%v", i/3, key, added, existed)
					}
					current.model[key] = value
				} else {
					var removed bool
					if current.node != nil {
						if owner == nil {
							current.node, removed = current.node.del(hash, 0, key)
						} else {
							current.node, removed = current.node.tdel(owner, hash, 0, key)
						}
					}
					if removed != existed {
						t.Fatalf("step %d: delete key %d removed=%v, previously present=%v", i/3, key, removed, existed)
					}
					delete(current.model, key)
				}
				if owner == nil {
					publish()
				}
			case 4:
				publish()
			case 5, 6:
				publish()
				base := history[int(data[i+1])%len(history)]
				current = version{node: base.node, model: maps.Clone(base.model)}
				if op == 5 {
					owner = ownership.New()
				}
			case 7:
				// Read-only checkpoint, including during an active batch.
			}
			check(i / 3)
		}
		publish()
		check(len(data) / 3)
	})
}

// assertCollisionModel checks content independently of node shape and iteration order.
func assertCollisionModel(t *testing.T, step, snapshot int, n node[int, int], want map[int]int) {
	t.Helper()
	for key := range 32 {
		var got int
		var found bool
		if n != nil {
			got, found = n.get(uint64(key%4), 0, key)
		}
		value, present := want[key]
		if got != value || found != present {
			t.Fatalf("step %d snapshot %d: get(%d)=(%d,%v), want (%d,%v)", step, snapshot, key, got, found, value, present)
		}
	}
	seen := make(map[int]int)
	if n != nil {
		if !n.forEach(func(key, value int) bool {
			if _, duplicate := seen[key]; duplicate {
				t.Fatalf("step %d snapshot %d: duplicate key %d", step, snapshot, key)
			}
			seen[key] = value
			return true
		}) {
			t.Fatalf("step %d snapshot %d: complete traversal stopped", step, snapshot)
		}
		pulls := 0
		if n.forEach(func(int, int) bool {
			pulls++
			return false
		}) || pulls != 1 {
			t.Fatalf("step %d snapshot %d: early stop pulled %d entries", step, snapshot, pulls)
		}
	}
	if !maps.Equal(seen, want) {
		t.Fatalf("step %d snapshot %d: entries=%v, want %v", step, snapshot, seen, want)
	}
}
