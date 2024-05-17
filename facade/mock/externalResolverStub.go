package mock

import "github.com/numbatx/gn-numbat/node/external"

type ExternalResolverStub struct {
	RecentNotarizedBlocksCalled func(maxShardHeadersNum int) ([]*external.BlockHeader, error)
	RetrieveShardBlockCalled    func(blockHash []byte) (*external.ShardBlockInfo, error)
}

func (ers *ExternalResolverStub) RecentNotarizedBlocks(maxShardHeadersNum int) ([]*external.BlockHeader, error) {
	return ers.RecentNotarizedBlocksCalled(maxShardHeadersNum)
}

func (ers *ExternalResolverStub) RetrieveShardBlock(blockHash []byte) (*external.ShardBlockInfo, error) {
	return ers.RetrieveShardBlockCalled(blockHash)
}
