package interceptors

import (
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

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	hdrInterceptorBase *HeaderInterceptorBase
	headers            storage.Cacher
	headersNonces      dataRetriever.Uint64Cacher
	shardCoordinator   sharding.Coordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*HeaderInterceptor, error) {

	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	hdrBaseInterceptor, err := NewHeaderInterceptorBase(
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

	hdrIntercept := &HeaderInterceptor{
		hdrInterceptorBase: hdrBaseInterceptor,
		headers:            headers,
		headersNonces:      headersNonces,
		shardCoordinator:   shardCoordinator,
	}

	return hdrIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (hi *HeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := hi.hdrInterceptorBase.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	go hi.processHeader(hdrIntercepted)

	return nil
}

func (hi *HeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	err := hi.hdrInterceptorBase.storer.Has(hdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return
	}
	if !hi.hdrInterceptorBase.CheckHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	hi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())
	hi.headersNonces.HasOrAdd(hdrIntercepted.GetHeader().Nonce, hdrIntercepted.Hash())
}
