package blockchain

import (
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/data/block"
	"github.com/numbatx/gn-numbat/storage"
)

// MetaChain holds the block information for the beacon chain
//
// The MetaChain also holds pointers to the Genesis block, the current block
// the height of the local chain and the perceived height of the chain in the network.
type MetaChain struct {
	GenesisBlock     *block.MetaBlock // Genesys Block pointer
	genesisBlockHash []byte           // Genesis Block hash
	CurrentBlock     *block.MetaBlock // Current Block pointer
	currentBlockHash []byte           // Current Block hash
	localHeight      int64            // Height of the local chain
	networkHeight    int64            // Percieved height of the network chain
	badBlocks        storage.Cacher   // Bad blocks cache
}

// NewMetaChain will initialize a new metachain instance
func NewMetaChain(
	badBlocksCache storage.Cacher,
) (*MetaChain, error) {
	if badBlocksCache == nil {
		return nil, ErrBadBlocksCacheNil
	}

	return &MetaChain{
		badBlocks: badBlocksCache,
	}, nil
}

// GetGenesisHeader returns the genesis block header pointer
func (mc *MetaChain) GetGenesisHeader() data.HeaderHandler {
	if mc.GenesisBlock == nil {
		return nil
	}
	return mc.GenesisBlock
}

// SetGenesisHeader returns the genesis block header pointer
func (mc *MetaChain) SetGenesisHeader(header data.HeaderHandler) error {
	if header == nil {
		mc.GenesisBlock = nil
		return nil
	}

	genBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}
	mc.GenesisBlock = genBlock
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (mc *MetaChain) GetGenesisHeaderHash() []byte {
	return mc.genesisBlockHash
}

// SetGenesisHeaderHash returns the genesis block header hash
func (mc *MetaChain) SetGenesisHeaderHash(headerHash []byte) {
	mc.genesisBlockHash = headerHash
}

// GetCurrentBlockHeader returns current block header pointer
func (mc *MetaChain) GetCurrentBlockHeader() data.HeaderHandler {
	if mc.CurrentBlock == nil {
		return nil
	}
	return mc.CurrentBlock
}

// SetCurrentBlockHeader sets current block header pointer
func (mc *MetaChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if header == nil {
		mc.CurrentBlock = nil
		return nil
	}

	currHead, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}

	mc.CurrentBlock = currHead

	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) GetCurrentBlockHeaderHash() []byte {
	return mc.currentBlockHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) SetCurrentBlockHeaderHash(hash []byte) {
	mc.currentBlockHash = hash
}

// GetCurrentBlockBody returns the block body pointer
func (mc *MetaChain) GetCurrentBlockBody() data.BodyHandler {
	return nil
}

// SetCurrentBlockBody sets the block body pointer
func (mc *MetaChain) SetCurrentBlockBody(body data.BodyHandler) error {
	// not needed to be implemented in metachain.
	return nil
}

// GetLocalHeight returns the height of the local chain
func (mc *MetaChain) GetLocalHeight() int64 {
	return mc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (mc *MetaChain) SetLocalHeight(height int64) {
	mc.localHeight = height
}

// GetNetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) GetNetworkHeight() int64 {
	return mc.networkHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) SetNetworkHeight(height int64) {
	mc.networkHeight = height
}

// HasBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (mc *MetaChain) HasBadBlock(blockHash []byte) bool {
	return mc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (mc *MetaChain) PutBadBlock(blockHash []byte) {
	mc.badBlocks.Put(blockHash, struct{}{})
}
