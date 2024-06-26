package state

import (
	"math/rand"
	"sync"
	"time"

	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/hashing/sha256"
	"github.com/numbatx/gn-numbat/storage"
	"github.com/numbatx/gn-numbat/storage/memorydb"
)

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func createDummyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())

	mutex.Lock()
	r.Read(buff)
	mutex.Unlock()

	return state.NewAddress(buff)
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	mutex.Lock()
	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[r.Int()%16]
	}
	mutex.Unlock()

	return string(buff)
}
