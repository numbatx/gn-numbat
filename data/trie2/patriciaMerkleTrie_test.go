package trie2_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/numbatx/gn-numbat/data/trie2"
	"github.com/numbatx/gn-numbat/hashing/keccak"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func initTrieMultipleValues(nr int) (trie2.Trie, [][]byte) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	var values [][]byte
	hsh := keccak.Keccak{}

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(fmt.Sprintf("%c", i)))
		tr.Update(values[i], values[i])
	}

	return tr, values

}

func initTrie() trie2.Trie {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func TestNewTrieWithNilDB(t *testing.T) {
	tr, err := trie2.NewTrie(nil, marshal.JsonMarshalizer{}, keccak.Keccak{})

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	db, _ := memorydb.New()
	tr, err := trie2.NewTrie(db, nil, keccak.Keccak{})

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	db, _ := memorydb.New()
	tr, err := trie2.NewTrie(db, marshal.JsonMarshalizer{}, nil)

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	tr, val := initTrieMultipleValues(10000)

	for i := range val {
		v, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_GetEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, trie2.ErrNilNode, err)
	assert.Nil(t, val)
}

func TestPatriciaMerkleTree_Update(t *testing.T) {
	tr := initTrie()

	newVal := []byte("doge")
	tr.Update([]byte("dog"), newVal)

	val, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	tr := initTrie()
	var empty []byte

	tr.Update([]byte("doe"), []byte{})

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	tr := initTrie()

	tr.Update([]byte("does"), []byte("this"))

	v, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	tr := initTrie()
	var empty []byte

	tr.Delete([]byte("doe"))

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_DeleteEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	err := tr.Delete([]byte("dog"))
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	tr := initTrie()

	root, err := tr.Root()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_NilRoot(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	root, err := tr.Root()
	assert.Equal(t, trie2.ErrNilNode, err)
	assert.Nil(t, root)
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	tr := initTrie()

	proof, err := tr.Prove([]byte("dog"))
	assert.Nil(t, err)
	ok, _ := tr.VerifyProof(proof, []byte("dog"))
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveCollapsedTrie(t *testing.T) {
	tr := initTrie()
	tr.Commit()

	proof, err := tr.Prove([]byte("dog"))
	assert.Nil(t, err)
	ok, _ := tr.VerifyProof(proof, []byte("dog"))
	assert.True(t, ok)
}

func TestPatriciaMerkleTree_ProveOnEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	proof, err := tr.Prove([]byte("dog"))
	assert.Nil(t, proof)
	assert.Equal(t, trie2.ErrNilNode, err)
}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	tr, val := initTrieMultipleValues(50)

	for i := range val {
		proof, _ := tr.Prove(val[i])

		ok, err := tr.VerifyProof(proof, val[i])
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proof, []byte("dog"+strconv.Itoa(i)))
		assert.Nil(t, err)
		assert.False(t, ok)
	}

}

func TestPatriciaMerkleTree_VerifyProofNilProofs(t *testing.T) {
	tr := initTrie()

	ok, err := tr.VerifyProof(nil, []byte("dog"))
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_VerifyProofEmptyProofs(t *testing.T) {
	tr := initTrie()

	ok, err := tr.VerifyProof([][]byte{}, []byte("dog"))
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	tr := initTrie()
	root1, _ := tr.Root()

	tr.Update([]byte("dodge"), []byte("viper"))
	root2, _ := tr.Root()

	tr.Delete([]byte("dodge"))
	root3, _ := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTree_Commit(t *testing.T) {
	tr := initTrie()

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitAfterCommit(t *testing.T) {
	tr := initTrie()

	tr.Commit()
	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitEmptyRoot(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	err := tr.Commit()
	assert.Equal(t, trie2.ErrNilNode, err)
}

func TestPatriciaMerkleTree_GetAfterCommit(t *testing.T) {
	tr := initTrie()

	err := tr.Commit()
	assert.Nil(t, err)

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, []byte("puppy"), val)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_InsertAfterCommit(t *testing.T) {
	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	tr1.Update([]byte("doge"), []byte("coin"))
	tr2.Update([]byte("doge"), []byte("coin"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)

}

func TestPatriciaMerkleTree_DeleteAfterCommit(t *testing.T) {
	tr1 := initTrie()
	tr2 := initTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	tr1.Delete([]byte("dogglesworth"))
	tr2.Delete([]byte("dogglesworth"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)
}

func emptyTrie() trie2.Trie {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})
	return tr
}

func BenchmarkPatriciaMerkleTree_Insert(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	nrValuesNotInTrie := 9000000
	values := make([][]byte, nrValuesNotInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		val := hsh.Compute(strconv.Itoa(i))
		tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_InsertCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 1000000
	nrValuesNotInTrie := 9000000
	values := make([][]byte, nrValuesNotInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		val := hsh.Compute(strconv.Itoa(i))
		tr.Update(val, val)
	}
	for i := 0; i < nrValuesNotInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i + nrValuesInTrie))
	}
	tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Update(values[i%nrValuesNotInTrie], values[i%nrValuesNotInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Delete(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_DeleteCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Get(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_GetCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}
	tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_Prove(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 3000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Prove(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_ProveCollapsedTrie(b *testing.B) {
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrValuesInTrie := 2000000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}
	tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Prove(values[i%nrValuesInTrie])
	}
}

func BenchmarkPatriciaMerkleTree_VerifyProof(b *testing.B) {
	var err error
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrProofs := 10
	proofs := make([][][]byte, nrProofs)

	nrValuesInTrie := 100000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}
	for i := 0; i < nrProofs; i++ {
		proofs[i], err = tr.Prove(values[i])
		assert.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.VerifyProof(proofs[i%nrProofs], values[i%nrProofs])
	}
}

func BenchmarkPatriciaMerkleTree_VerifyProofCollapsedTrie(b *testing.B) {
	var err error
	tr := emptyTrie()
	hsh := keccak.Keccak{}

	nrProofs := 10
	proofs := make([][][]byte, nrProofs)

	nrValuesInTrie := 100000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}
	for i := 0; i < nrProofs; i++ {
		proofs[i], err = tr.Prove(values[i])
		assert.Nil(b, err)
	}
	tr.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.VerifyProof(proofs[i%nrProofs], values[i%nrProofs])
	}
}

func BenchmarkPatriciaMerkleTree_Commit(b *testing.B) {
	nrValuesInTrie := 1000000
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hsh := keccak.Keccak{}
		tr := emptyTrie()
		for i := 0; i < nrValuesInTrie; i++ {
			hash := hsh.Compute(strconv.Itoa(i))
			tr.Update(hash, hash)
		}
		b.StartTimer()

		tr.Commit()
	}
}
