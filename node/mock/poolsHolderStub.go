package mock

import (
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/storage"
)

type PoolsHolderStub struct {
	HeadersCalled           func() storage.Cacher
	HeadersNoncesCalled     func() dataRetriever.Uint64Cacher
	PeerChangesBlocksCalled func() storage.Cacher
	TransactionsCalled      func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled        func() storage.Cacher
	MetaBlocksCalled        func() storage.Cacher
	MetaHeadersNoncesCalled func() dataRetriever.Uint64Cacher
}

func (phs *PoolsHolderStub) Headers() storage.Cacher {
	return phs.HeadersCalled()
}

func (phs *PoolsHolderStub) HeadersNonces() dataRetriever.Uint64Cacher {
	return phs.HeadersNoncesCalled()
}

func (phs *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	return phs.PeerChangesBlocksCalled()
}

func (phs *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.TransactionsCalled()
}

func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return phs.MiniBlocksCalled()
}

func (phs *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return phs.MetaBlocksCalled()
}

func (phs *PoolsHolderStub) MetaHeadersNonces() dataRetriever.Uint64Cacher {
	return phs.MetaHeadersNoncesCalled()
}
