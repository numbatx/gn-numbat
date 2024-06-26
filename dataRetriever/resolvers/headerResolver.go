package resolvers

import (
	"fmt"

	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/data/typeConverters"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/storage"
)

var log = logger.DefaultLogger()

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	*HeaderResolverBase
	hdrNonces      dataRetriever.Uint64Cacher
	headers        storage.Cacher
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(
	senderResolver dataRetriever.TopicResolverSender,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64Cacher,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	nonceConverter typeConverters.Uint64ByteSliceConverter,
) (*HeaderResolver, error) {

	if senderResolver == nil {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if headers == nil {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if headersNonces == nil {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}
	if nonceConverter == nil {
		return nil, dataRetriever.ErrNilNonceConverter
	}
	hdrResolverBase, err := NewHeaderResolverBase(
		senderResolver,
		headers,
		hdrStorage,
		marshalizer,
	)
	if err != nil {
		return nil, err
	}

	hdrResolver := &HeaderResolver{
		hdrNonces:          headersNonces,
		headers:            headers,
		nonceConverter:     nonceConverter,
		HeaderResolverBase: hdrResolverBase,
	}

	return hdrResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (hdrRes *HeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
	rd, err := hdrRes.ParseReceivedMessage(message)
	if err != nil {
		return err
	}
	var buff []byte

	switch rd.Type {
	case dataRetriever.HashType:
		buff, err = hdrRes.ResolveHeaderFromHash(rd.Value)
	case dataRetriever.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd.Value)
	default:
		return dataRetriever.ErrResolveTypeUnknown
	}
	if err != nil {
		return err
	}
	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return hdrRes.Send(buff, message.Peer())
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(key []byte) ([]byte, error) {
	//key is now an encoded nonce (uint64)

	//Step 1. decode the nonce from the key
	nonce, err := hdrRes.nonceConverter.ToUint64(key)
	if err != nil {
		return nil, dataRetriever.ErrInvalidNonceByteSlice
	}

	//Step 2. search the nonce-key pair
	hash, _ := hdrRes.hdrNonces.Get(nonce)
	if hash == nil {
		return nil, nil
	}

	//Step 3. search header by key (hash)
	value, ok := hdrRes.headers.Peek(hash)
	if !ok {
		return hdrRes.hdrStorage.Get(hash)
	}

	//since there might be multiple entries, it shall return the first one that it finds
	buff, err := hdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestDataFromNonce requests a header from other peers having input the hdr nonce
func (hdrRes *HeaderResolver) RequestDataFromNonce(nonce uint64) error {
	return hdrRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.NonceType,
		Value: hdrRes.nonceConverter.ToByteSlice(nonce),
	})
}
