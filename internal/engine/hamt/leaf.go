package hamt

import "go.alvadorn.org/functional/internal/core/ownership"

// leaf stores a single key/value at a known full hash.
type leaf[K comparable, V any] struct {
	owner *ownership.Owner // nil for published path-copy nodes
	hash  uint64
	key   K
	val   V
}

func (l *leaf[K, V]) get(hash uint64, _ uint, key K) (V, bool) {
	if l.hash == hash && l.key == key {
		return l.val, true
	}
	var zero V
	return zero, false
}

func (l *leaf[K, V]) set(hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	if l.hash == hash && l.key == key {
		// Replace key and value; still allocate so the old leaf stays immutable.
		return &leaf[K, V]{hash: hash, key: key, val: val}, false
	}
	if l.hash == hash {
		return &collision[K, V]{
			hash: hash,
			entries: []entry[K, V]{
				{key: l.key, val: l.val},
				{key: key, val: val},
			},
		}, true
	}
	n2 := &leaf[K, V]{hash: hash, key: key, val: val}
	return merge(shift, l.hash, l, hash, n2, nil), true
}

func (l *leaf[K, V]) del(hash uint64, _ uint, key K) (node[K, V], bool) {
	if l.hash == hash && l.key == key {
		return nil, true
	}
	return l, false
}

func (l *leaf[K, V]) tset(owner *ownership.Owner, hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	if l.hash == hash && l.key == key {
		n := ensureEditableLeaf(l, owner)
		n.key, n.val = key, val
		return n, false
	}
	if l.hash == hash {
		// Convert to collision under owner (old leaf exclusive or discarded).
		n := ensureEditableLeaf(l, owner)
		return &collision[K, V]{
			owner: owner,
			hash:  hash,
			entries: []entry[K, V]{
				{key: n.key, val: n.val},
				{key: key, val: val},
			},
		}, true
	}
	n := ensureEditableLeaf(l, owner)
	n2 := &leaf[K, V]{owner: owner, hash: hash, key: key, val: val}
	return merge(shift, n.hash, n, hash, n2, owner), true
}

func (l *leaf[K, V]) tdel(_ *ownership.Owner, hash uint64, _ uint, key K) (node[K, V], bool) {
	if l.hash == hash && l.key == key {
		return nil, true
	}
	return l, false
}

func (l *leaf[K, V]) forEach(yield func(K, V) bool) bool {
	return yield(l.key, l.val)
}
