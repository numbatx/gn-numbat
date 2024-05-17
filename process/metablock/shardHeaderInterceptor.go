package metablock

import (
	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/block"
	"github.com/numbatx/gn-numbat/process/block/interceptors"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/numbatx/gn-numbat/storage"
)

var log = logger.DefaultLogger()

// ShardHeaderInterceptor represents an interceptor used for shard block headers by metachain nodes
type ShardHeaderInterceptor struct {
	hdrInterceptorBase *interceptors.HeaderInterceptorBase
	headers            storage.Cacher
	storer             storage.Storer
}

// NewShardHeaderInterceptor hooks a new interceptor for shard block headers by metachain nodes
// Fetched block headers will be placed in a data pool
func NewShardHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*ShardHeaderInterceptor, error) {

	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	hdrBaseInterceptor, err := interceptors.NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multiSigVerifier,
		hasher,
		shardCoordinator,
		chronologyValidator,
	)
	if err != nil {
		return nil, err
	}

	return &ShardHeaderInterceptor{
		hdrInterceptorBase: hdrBaseInterceptor,
		headers:            headers,
		storer:             storer,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (shi *ShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := shi.hdrInterceptorBase.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	go shi.processHeader(hdrIntercepted)

	return nil
}

func (shi *ShardHeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	err := shi.storer.Has(hdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return
	}
	if !shi.hdrInterceptorBase.CheckHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	shi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())
}
