package hamt

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

// entry is one key/value pair inside a collision bucket.
type entry[K comparable, V any] struct {
	key K
	val V
}

// collision holds multiple keys that share the same 64-bit hash.
type collision[K comparable, V any] struct {
	owner   *ownership.Owner // nil for published path-copy nodes
	hash    uint64
	entries []entry[K, V]
}

func (c *collision[K, V]) get(hash uint64, _ uint, key K) (V, bool) {
	if c.hash != hash {
		var zero V
		return zero, false
	}
	for i := range c.entries {
		if c.entries[i].key == key {
			return c.entries[i].val, true
		}
	}
	var zero V
	return zero, false
}

func (c *collision[K, V]) set(hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	if c.hash != hash {
		// Divergent hash at this depth: push both sides into a bitmap path.
		n2 := &leaf[K, V]{hash: hash, key: key, val: val}
		return merge(shift, c.hash, c, hash, n2, nil), true
	}
	for i := range c.entries {
		if c.entries[i].key == key {
			next := make([]entry[K, V], len(c.entries))
			copy(next, c.entries)
			next[i] = entry[K, V]{key: key, val: val}
			return &collision[K, V]{hash: c.hash, entries: next}, false
		}
	}
	next := make([]entry[K, V], len(c.entries)+1)
	copy(next, c.entries)
	next[len(c.entries)] = entry[K, V]{key: key, val: val}
	return &collision[K, V]{hash: c.hash, entries: next}, true
}

func (c *collision[K, V]) del(hash uint64, _ uint, key K) (node[K, V], bool) {
	if c.hash != hash {
		return c, false
	}
	idx := -1
	for i := range c.entries {
		if c.entries[i].key == key {
			idx = i
			break
		}
	}
	if idx < 0 {
		return c, false
	}
	if len(c.entries) == 1 {
		return nil, true
	}
	if len(c.entries) == 2 {
		// Collapse back to a single leaf.
		keep := c.entries[1-idx]
		return &leaf[K, V]{hash: c.hash, key: keep.key, val: keep.val}, true
	}
	next := make([]entry[K, V], 0, len(c.entries)-1)
	next = append(next, c.entries[:idx]...)
	next = append(next, c.entries[idx+1:]...)
	return &collision[K, V]{hash: c.hash, entries: next}, true
}

func (c *collision[K, V]) tset(owner *ownership.Owner, hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	if c.hash != hash {
		n := ensureEditableCollision(c, owner)
		n2 := &leaf[K, V]{owner: owner, hash: hash, key: key, val: val}
		return merge(shift, n.hash, n, hash, n2, owner), true
	}
	for i := range c.entries {
		if c.entries[i].key == key {
			n := ensureEditableCollision(c, owner)
			n.entries[i] = entry[K, V]{key: key, val: val}
			return n, false
		}
	}
	n := ensureEditableCollision(c, owner)
	n.entries = append(n.entries, entry[K, V]{key: key, val: val})
	return n, true
}

func (c *collision[K, V]) tdel(owner *ownership.Owner, hash uint64, _ uint, key K) (node[K, V], bool) {
	if c.hash != hash {
		return c, false
	}
	idx := -1
	for i := range c.entries {
		if c.entries[i].key == key {
			idx = i
			break
		}
	}
	if idx < 0 {
		return c, false
	}
	if len(c.entries) == 1 {
		return nil, true
	}
	if len(c.entries) == 2 {
		keep := c.entries[1-idx]
		return &leaf[K, V]{owner: owner, hash: c.hash, key: keep.key, val: keep.val}, true
	}
	n := ensureEditableCollision(c, owner)
	// idx still valid: entries layout unchanged by shallow clone.
	next := make([]entry[K, V], 0, len(n.entries)-1)
	next = append(next, n.entries[:idx]...)
	next = append(next, n.entries[idx+1:]...)
	n.entries = next
	return n, true
}

func (c *collision[K, V]) forEach(yield func(K, V) bool) bool {
	for i := range c.entries {
		if !yield(c.entries[i].key, c.entries[i].val) {
			return false
		}
	}
	return true
}
