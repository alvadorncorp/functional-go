package keyhash_test

import (
	"hash/maphash"
	"testing"

	"go.alvadorn.org/functional/internal/engine/keyhash"
)

func TestHashDeterministicSameSeed(t *testing.T) {
	t.Parallel()

	seed := maphash.MakeSeed()
	const key = "functional-keyhash"

	h1 := keyhash.Hash(seed, key)
	h2 := keyhash.Hash(seed, key)
	if h1 != h2 {
		t.Fatalf("same seed+key must be deterministic: %d != %d", h1, h2)
	}
}

func TestHashDifferentKeysDiffer(t *testing.T) {
	t.Parallel()

	seed := maphash.MakeSeed()
	if keyhash.Hash(seed, "a") == keyhash.Hash(seed, "b") {
		// Cryptographic collision on short distinct strings is vanishingly rare;
		// treat equality as a test environment failure worth flagging.
		t.Fatal("distinct keys under one seed produced equal hashes")
	}
}

func TestHashDifferentSeedsMayDiffer(t *testing.T) {
	t.Parallel()

	// Seeds are independent; hashes for the same key are not required to
	// differ, but with high probability they do. Retry a few pairs so a
	// single accidental collision does not flake the suite.
	const key = 42
	for range 8 {
		s1 := maphash.MakeSeed()
		s2 := maphash.MakeSeed()
		if keyhash.Hash(s1, key) != keyhash.Hash(s2, key) {
			return
		}
	}
	t.Fatal("expected at least one seed pair to hash the same key differently")
}

func TestHashIntAndStruct(t *testing.T) {
	t.Parallel()

	seed := maphash.MakeSeed()
	type pair struct {
		A int
		B string
	}

	hInt1 := keyhash.Hash(seed, 7)
	hInt2 := keyhash.Hash(seed, 7)
	if hInt1 != hInt2 {
		t.Fatal("int hash not stable")
	}
	p := pair{A: 1, B: "x"}
	hStruct1 := keyhash.Hash(seed, p)
	hStruct2 := keyhash.Hash(seed, p)
	if hStruct1 != hStruct2 {
		t.Fatal("struct hash not stable")
	}
}
