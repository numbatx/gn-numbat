package data

// HeaderHandler defines getters and setters for header data holder
type HeaderHandler interface {
	GetNonce() uint64
	GetEpoch() uint32
	GetRound() uint32
	GetRootHash() []byte
	GetPrevHash() []byte
	GetPrevRandSeed() []byte
	GetRandSeed() []byte
	GetPubKeysBitmap() []byte
	GetSignature() []byte
	GetTimeStamp() uint64
	GetTxCount() uint32

	SetNonce(n uint64)
	SetEpoch(e uint32)
	SetRound(r uint32)
	SetTimeStamp(ts uint64)
	SetRootHash(rHash []byte)
	SetPrevHash(pvHash []byte)
	SetPrevRandSeed(pvRandSeed []byte)
	SetRandSeed(randSeed []byte)
	SetPubKeysBitmap(pkbm []byte)
	SetSignature(sg []byte)
	SetTxCount(txCount uint32)

	GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32
	GetMiniBlockProcessed(hash []byte) bool
	SetMiniBlockProcessed(hash []byte, processed bool)
}

// BodyHandler interface for a block body
type BodyHandler interface {
	// IntegrityAndValidity checks the integrity and validity of the block
	IntegrityAndValidity() error
}

// ChainHandler is the interface defining the functionality a blockchain should implement
type ChainHandler interface {
	GetGenesisHeader() HeaderHandler
	SetGenesisHeader(gb HeaderHandler) error
	GetGenesisHeaderHash() []byte
	SetGenesisHeaderHash(hash []byte)
	GetCurrentBlockHeader() HeaderHandler
	SetCurrentBlockHeader(bh HeaderHandler) error
	GetCurrentBlockHeaderHash() []byte
	SetCurrentBlockHeaderHash(hash []byte)
	GetCurrentBlockBody() BodyHandler
	SetCurrentBlockBody(body BodyHandler) error
	GetLocalHeight() int64
	SetLocalHeight(height int64)
	GetNetworkHeight() int64
	SetNetworkHeight(height int64)
	HasBadBlock(blockHash []byte) bool
	PutBadBlock(blockHash []byte)
}
