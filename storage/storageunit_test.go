package storage_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/hashing/blake2b"
	"github.com/numbatx/gn-numbat/hashing/fnv"
	"github.com/numbatx/gn-numbat/hashing/keccak"

	"github.com/numbatx/gn-numbat/storage"
	"github.com/numbatx/gn-numbat/storage/bloom"
	"github.com/numbatx/gn-numbat/storage/leveldb"
	"github.com/numbatx/gn-numbat/storage/lrucache"
	"github.com/numbatx/gn-numbat/storage/memorydb"

	"github.com/stretchr/testify/assert"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func initStorageUnitWithBloomFilter(t *testing.T, cSize int) *storage.Unit {
	mdb, err1 := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "failed creating db: %s", err1)
	assert.Nil(t, err2, "no error expected but got %s", err2)

	sUnit, err := storage.NewStorageUnitWithBloomFilter(cache, mdb, bf)

	assert.Nil(t, err, "failed to create storage unit")

	return sUnit
}

func initStorageUnitWithNilBloomFilter(t *testing.T, cSize int) *storage.Unit {
	mdb, err1 := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)

	assert.Nil(t, err1, "failed creating db: %s", err1)
	assert.Nil(t, err2, "no error expected but got %s", err2)

	sUnit, err := storage.NewStorageUnit(cache, mdb)

	assert.Nil(t, err, "failed to create storage unit")

	return sUnit
}

func TestStorageUnitNilPersister(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "no error expected but got %s", err1)

	_, err := storage.NewStorageUnitWithBloomFilter(cache, nil, bf)

	assert.NotNil(t, err, "expected failure")
}

func TestStorageUnitNilCacher(t *testing.T) {
	mdb, err1 := memorydb.New()
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "failed creating db")

	_, err1 = storage.NewStorageUnitWithBloomFilter(nil, mdb, bf)

	assert.NotNil(t, err1, "expected failure")
}

func TestStorageUnitNilBloomFilter(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	mdb, err2 := memorydb.New()

	assert.Nil(t, err1, "no error expected but got %s", err1)
	assert.Nil(t, err2, "failed creating db")

	_, err := storage.NewStorageUnit(cache, mdb)

	assert.Nil(t, err, "did not expect failure")
}

func TestStorageUnit_NilBloomFilterShouldErr(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	mdb, err2 := memorydb.New()

	assert.Nil(t, err1, "no error expected but got %s", err1)
	assert.Nil(t, err2, "failed creating db")

	sUnit, err := storage.NewStorageUnitWithBloomFilter(cache, mdb, nil)

	assert.Nil(t, sUnit)
	assert.Equal(t, "expected not nil bloom filter", err.Error())
}

func TestPutNotPresent(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutNotPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutNotPresentCache(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "expected to find key %s, but not found", key)
}

func TestPutPresent(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	// put again same value, no error expected
	err = s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	// put again same value, no error expected
	err = s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key3")
	s := initStorageUnitWithBloomFilter(t, 10)
	v, err := s.Get(key)

	assert.NotNil(t, err, "expected to find no value, but found %s", v)
}

func TestGetNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key3")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	v, err := s.Get(key)

	assert.NotNil(t, err, "expected to find no value, but found %s", v)
}

func TestGetNotPresentCache(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value4")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key5"), []byte("value4")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key6")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestHasNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key6")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestHasNotPresentCache(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddNotPresent(t *testing.T) {
	key, val := []byte("key9"), []byte("value9")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.HasOrAdd(key, val)

	assert.Nil(t, err)
	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddNotPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key9"), []byte("value9")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.HasOrAdd(key, val)

	assert.Nil(t, err)
	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddNotPresentCache(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	s.ClearCache()

	err = s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	s.ClearCache()

	err = s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddPresent(t *testing.T) {
	key, val := []byte("key11"), []byte("value11")
	s := initStorageUnitWithBloomFilter(t, 10)
	_ = s.Put(key, val)

	err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasOrAddPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key11"), []byte("value11")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	_ = s.Put(key, val)

	err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresent(t *testing.T) {
	key := []byte("key12")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key12")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresentCache(t *testing.T) {
	key, val := []byte("key13"), []byte("value13")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	s.ClearCache()

	err = s.Remove(key)
	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestDeleteNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key13"), []byte("value13")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	s.ClearCache()

	err = s.Remove(key)
	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestDeletePresent(t *testing.T) {
	key, val := []byte("key14"), []byte("value14")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestDeletePresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key14"), []byte("value14")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Key not found")
}

func TestClearCacheNotAffectPersist(t *testing.T) {
	key, val := []byte("key15"), []byte("value15")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")
	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestDestroyUnitNoError(t *testing.T) {
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.DestroyUnit()
	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestDestroyUnitWithNilBloomFilterNoError(t *testing.T) {
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.DestroyUnit()
	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestCreateCacheFromConfWrongType(t *testing.T) {

	cacher, err := storage.NewCache("NotLRU", 100, 1)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, cacher, "cacher expected to be nil, but got %s", cacher)
}

func TestCreateCacheFromConfOK(t *testing.T) {

	cacher, err := storage.NewCache(storage.LRUCache, 10, 1)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, cacher, "valid cacher expected but got nil")
}

func TestCreateDBFromConfWrongType(t *testing.T) {
	persister, err := storage.NewDB("NotLvlDB", "test")

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfWrongFileNameLvlDB(t *testing.T) {
	persister, err := storage.NewDB(storage.LvlDB, "")
	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfWrongFileNameBoltDB(t *testing.T) {
	persister, err := storage.NewDB(storage.BoltDB, "")
	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfWrongFileNameBadgerDB(t *testing.T) {
	persister, err := storage.NewDB(storage.BadgerDB, "")
	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfLvlDBOk(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	persister, err := storage.NewDB(storage.LvlDB, dir)
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, persister, "valid persister expected but got nil")

	err = persister.Destroy()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestCreateDBFromConfBoltDBOk(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	persister, err := storage.NewDB(storage.BoltDB, dir)
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, persister, "valid persister expected but got nil")

	err = persister.Destroy()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestCreateDBFromConfBadgerDBOk(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	persister, err := storage.NewDB(storage.BadgerDB, dir)
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, persister, "valid persister expected but got nil")

	err = persister.Destroy()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestCreateBloomFilterFromConfWrongSize(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, bf, "persister expected to be nil, but got %s", bf)
}

func TestCreateBloomFilterFromConfWrongHashFunc(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, bf, "persister expected to be nil, but got %s", bf)
}

func TestCreateBloomFilterFromConfOk(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, bf, "valid persister expected but got nil")
}

func TestNewStorageUnit_FromConfWrongCacheConfig(t *testing.T) {

	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: "NotLRU",
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnit_FromConfWrongDBConfig(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     "NotLvlDB",
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnit_FromConfLvlDBOk(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_FromConfBoltDBOk(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BoltDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_FromConfBadgerDBOk(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BadgerDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithBlankBloomFilterShouldWorkLvlDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.Nil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithBlankBloomFilterShouldWorkBoltDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BoltDB,
	}, storage.BloomConfig{})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.Nil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithBlankBloomFilterShouldWorkBadgerDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BadgerDB,
	}, storage.BloomConfig{})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.Nil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithConfigBloomFilterShouldCreateBloomFilterLvlDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.NotNil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithConfigBloomFilterShouldCreateBloomFilterBoltDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BoltDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.NotNil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithConfigBloomFilterShouldCreateBloomFilterBadgerDB(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BadgerDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	assert.NotNil(t, storer.GetBlomFilter())
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_WithInvalidConfigBloomFilterLvlDBShouldErr(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.HasherType("invalid"), storage.Fnv},
	})

	assert.Equal(t, "hash type not supported", err.Error())
	assert.Nil(t, storer)
}

func TestNewStorageUnit_WithInvalidConfigBloomFilterBoltDBShouldErr(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BoltDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.HasherType("invalid"), storage.Fnv},
	})

	assert.Equal(t, "hash type not supported", err.Error())
	assert.Nil(t, storer)
}

func TestNewStorageUnit_WithInvalidConfigBloomFilterBadgerDBShouldErr(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.BadgerDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.HasherType("invalid"), storage.Fnv},
	})

	assert.Equal(t, "hash type not supported", err.Error())
	assert.Nil(t, storer)
}

const (
	valuesInDb = 100000
	bfSize     = 100000
)

func initSUWithNilBloomFilter(cSize int) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := lrucache.NewCache(cSize)

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}

	sUnit, err := storage.NewStorageUnit(cache, ldb)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit

}

func initSUWithBloomFilter(cSize int, bfSize uint) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := lrucache.NewCache(cSize)
	bf, err3 := bloom.NewFilter(bfSize, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}})

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}
	if err3 != nil {
		fmt.Println(err3)
	}

	sUnit, err := storage.NewStorageUnitWithBloomFilter(cache, ldb, bf)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit
}

func BenchmarkStorageUnit_PutWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		err := s.Put([]byte(strconv.Itoa(nr)), []byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_PutWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		err := s.Put([]byte(strconv.Itoa(nr)), []byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetNotPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb) + valuesInDb
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetNotPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb) + valuesInDb
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}
