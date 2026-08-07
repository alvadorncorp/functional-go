package keyhash

import "hash/maphash"

// Hash returns a 64-bit hash of key under seed.
// It is a thin wrapper over maphash.Comparable: no type specialization,
// reflect, unsafe, or Hashable interface.
func Hash[K comparable](seed maphash.Seed, key K) uint64 {
	return maphash.Comparable(seed, key)
}
