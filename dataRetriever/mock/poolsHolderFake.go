package mock

import (
	"github.com/numbatx/gn-numbat/data/typeConverters/uint64ByteSlice"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/dataRetriever/dataPool"
	"github.com/numbatx/gn-numbat/dataRetriever/shardedData"
	"github.com/numbatx/gn-numbat/storage"
)

type PoolsHolderFake struct {
	transactions      dataRetriever.ShardedDataCacherNotifier
	headers           storage.Cacher
	metaBlocks        storage.Cacher
	hdrNonces         dataRetriever.Uint64Cacher
	miniBlocks        storage.Cacher
	peerChangesBlocks storage.Cacher
}

func NewPoolsHolderFake() *PoolsHolderFake {
	phf := &PoolsHolderFake{}
	phf.transactions, _ = shardedData.NewShardedData(storage.CacheConfig{Size: 10000, Type: storage.LRUCache})
	phf.headers, _ = storage.NewCache(storage.LRUCache, 10000, 1)
	phf.metaBlocks, _ = storage.NewCache(storage.LRUCache, 10000, 1)
	cacheHdrNonces, _ := storage.NewCache(storage.LRUCache, 10000, 1)
	phf.hdrNonces, _ = dataPool.NewNonceToHashCacher(
		cacheHdrNonces,
		uint64ByteSlice.NewBigEndianConverter(),
	)
	phf.miniBlocks, _ = storage.NewCache(storage.LRUCache, 10000, 1)
	phf.peerChangesBlocks, _ = storage.NewCache(storage.LRUCache, 10000, 1)
	return phf
}

func (phf *PoolsHolderFake) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phf.transactions
}

func (phf *PoolsHolderFake) Headers() storage.Cacher {
	return phf.headers
}

func (phf *PoolsHolderFake) HeadersNonces() dataRetriever.Uint64Cacher {
	return phf.hdrNonces
}

func (phf *PoolsHolderFake) MiniBlocks() storage.Cacher {
	return phf.miniBlocks
}

func (phf *PoolsHolderFake) PeerChangesBlocks() storage.Cacher {
	return phf.peerChangesBlocks
}

func (phf *PoolsHolderFake) MetaBlocks() storage.Cacher {
	return phf.metaBlocks
}
