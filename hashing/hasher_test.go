package hashing_test

import (
	"testing"

	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/hashing/blake2b"
	"github.com/numbatx/gn-numbat/hashing/fnv"
	"github.com/numbatx/gn-numbat/hashing/keccak"
	"github.com/numbatx/gn-numbat/hashing/sha256"
	"github.com/stretchr/testify/assert"
)

func TestSha256(t *testing.T) {
	Suite(t, sha256.Sha256{})
}

func TestBlake2b(t *testing.T) {
	Suite(t, blake2b.Blake2b{})
}

func TestKeccak(t *testing.T) {
	Suite(t, keccak.Keccak{})
}

func TestFnv(t *testing.T) {
	Suite(t, fnv.Fnv{})
}

func Suite(t *testing.T, h hashing.Hasher) {
	TestingCalculateHash(t, h)
	TestingCalculateEmptyHash(t, h)
	TestingNilReturn(t, h)

}

func TestingCalculateHash(t *testing.T, h hashing.Hasher) {

	h1 := h.Compute("a")
	h2 := h.Compute("b")

	assert.NotEqual(t, h1, h2)

}

func TestingCalculateEmptyHash(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("")
	h2 := h.EmptyHash()

	assert.Equal(t, h1, h2)

}

func TestingNilReturn(t *testing.T, h hashing.Hasher) {
	h1 := h.Compute("a")
	assert.NotNil(t, h1)
}
