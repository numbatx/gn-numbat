package interceptors

import (
	"github.com/numbatx/gn-numbat/core/statistics"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/block"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/numbatx/gn-numbat/storage"
)

// MetachainHeaderInterceptor represents an interceptor used for metachain block headers
type MetachainHeaderInterceptor struct {
	*messageChecker
	marshalizer            marshal.Marshalizer
	metachainHeaders       storage.Cacher
	metachainHeadersNonces dataRetriever.Uint64Cacher
	tpsBenchmark           *statistics.TpsBenchmark
	storer                 storage.Storer
	multiSigVerifier       crypto.MultiSigVerifier
	hasher                 hashing.Hasher
	shardCoordinator       sharding.Coordinator
	chronologyValidator    process.ChronologyValidator
}

// NewMetachainHeaderInterceptor hooks a new interceptor for metachain block headers
// Fetched metachain block headers will be placed in a data pool
func NewMetachainHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	metachainHeaders storage.Cacher,
	metachainHeadersNonces dataRetriever.Uint64Cacher,
	tpsBenchmark *statistics.TpsBenchmark,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*MetachainHeaderInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if metachainHeaders == nil {
		return nil, process.ErrNilMetachainHeadersDataPool
	}
	if metachainHeadersNonces == nil {
		return nil, process.ErrNilMetachainHeadersNoncesDataPool
	}
	if storer == nil {
		return nil, process.ErrNilMetachainHeadersStorage
	}
	if multiSigVerifier == nil {
		return nil, process.ErrNilMultiSigVerifier
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if chronologyValidator == nil {
		return nil, process.ErrNilChronologyValidator
	}

	return &MetachainHeaderInterceptor{
		messageChecker:         &messageChecker{},
		marshalizer:            marshalizer,
		metachainHeaders:       metachainHeaders,
		tpsBenchmark:           tpsBenchmark,
		storer:                 storer,
		multiSigVerifier:       multiSigVerifier,
		hasher:                 hasher,
		shardCoordinator:       shardCoordinator,
		chronologyValidator:    chronologyValidator,
		metachainHeadersNonces: metachainHeadersNonces,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mhi *MetachainHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := mhi.checkMessage(message)
	if err != nil {
		return err
	}

	metaHdrIntercepted := block.NewInterceptedMetaHeader(mhi.multiSigVerifier, mhi.chronologyValidator)
	err = mhi.marshalizer.Unmarshal(metaHdrIntercepted, message.Data())
	if err != nil {
		return err
	}

	hashWithSig := mhi.hasher.Compute(string(message.Data()))
	metaHdrIntercepted.SetHash(hashWithSig)

	err = metaHdrIntercepted.IntegrityAndValidity(mhi.shardCoordinator)
	if err != nil {
		return err
	}

	err = metaHdrIntercepted.VerifySig()
	if err != nil {
		return err
	}

	if mhi.tpsBenchmark != nil {
		mhi.tpsBenchmark.Update(metaHdrIntercepted.GetMetaHeader())
	}

	go mhi.processMetaHeader(metaHdrIntercepted)

	return nil
}

func (mhi *MetachainHeaderInterceptor) processMetaHeader(metaHdrIntercepted *block.InterceptedMetaHeader) {
	err := mhi.storer.Has(metaHdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted meta block header already processed")
		return
	}

	mhi.metachainHeaders.HasOrAdd(metaHdrIntercepted.Hash(), metaHdrIntercepted.GetMetaHeader())
	mhi.metachainHeadersNonces.HasOrAdd(metaHdrIntercepted.Nonce, metaHdrIntercepted.Hash())
}
