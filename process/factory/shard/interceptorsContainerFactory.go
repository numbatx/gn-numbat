package shard

import (
	"github.com/numbatx/gn-numbat/core/statistics"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/block/interceptors"
	"github.com/numbatx/gn-numbat/process/factory"
	"github.com/numbatx/gn-numbat/process/factory/containers"
	"github.com/numbatx/gn-numbat/process/transaction"
	"github.com/numbatx/gn-numbat/sharding"
)

type interceptorsContainerFactory struct {
	shardCoordinator    sharding.Coordinator
	messenger           process.TopicHandler
	store               dataRetriever.StorageService
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	keyGen              crypto.KeyGenerator
	singleSigner        crypto.SingleSigner
	multiSigner         crypto.MultiSigner
	dataPool            dataRetriever.PoolsHolder
	addrConverter       state.AddressConverter
	chronologyValidator process.ChronologyValidator
	tpsBenchmark        *statistics.TpsBenchmark
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	chronologyValidator process.ChronologyValidator,
	tpsBenchmark *statistics.TpsBenchmark,
) (*interceptorsContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if store == nil {
		return nil, process.ErrNilBlockChain
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if keyGen == nil {
		return nil, process.ErrNilKeyGen
	}
	if singleSigner == nil {
		return nil, process.ErrNilSingleSigner
	}
	if multiSigner == nil {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}
	if chronologyValidator == nil {
		return nil, process.ErrNilChronologyValidator
	}

	return &interceptorsContainerFactory{
		shardCoordinator:    shardCoordinator,
		messenger:           messenger,
		store:               store,
		marshalizer:         marshalizer,
		hasher:              hasher,
		keyGen:              keyGen,
		singleSigner:        singleSigner,
		multiSigner:         multiSigner,
		dataPool:            dataPool,
		addrConverter:       addrConverter,
		chronologyValidator: chronologyValidator,
		tpsBenchmark:        tpsBenchmark,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (icf *interceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	keys, interceptorSlice, err := icf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateHdrInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generatePeerChBlockBodyInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMetachainHeaderInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (icf *interceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := icf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, icf.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (icf *interceptorsContainerFactory) generateTxInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierTx
		interceptorSlice[int(idx)] = interceptor
	}

	//tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneTxInterceptor(identifier string) (process.Interceptor, error) {
	txStorer := icf.store.GetStorer(dataRetriever.TransactionUnit)

	interceptor, err := transaction.NewTxInterceptor(
		icf.marshalizer,
		icf.dataPool.Transactions(),
		txStorer,
		icf.addrConverter,
		icf.hasher,
		icf.singleSigner,
		icf.keyGen,
		icf.shardCoordinator)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Hdr interceptor

func (icf *interceptorsContainerFactory) generateHdrInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	//only one intrashard header topic
	identifierHdr := factory.HeadersTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	headerStorer := icf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	interceptor, err := interceptors.NewHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.Headers(),
		icf.dataPool.HeadersNonces(),
		headerStorer,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.chronologyValidator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

//------- MiniBlocks interceptors

func (icf *interceptorsContainerFactory) generateMiniBlocksInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorSlice[int(idx)] = interceptor
	}

	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icf.store.GetStorer(dataRetriever.MiniBlockUnit)

	interceptor, err := interceptors.NewTxBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.MiniBlocks(),
		txBlockBodyStorer,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- PeerChBlocks interceptor

func (icf *interceptorsContainerFactory) generatePeerChBlockBodyInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := factory.PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	peerBlockBodyStorer := icf.store.GetStorer(dataRetriever.PeerChangesUnit)

	interceptor, err := interceptors.NewPeerBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		icf.hasher,
		shardC,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierPeerCh, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []process.Interceptor{interceptor}, nil
}

//------- MetachainHeader interceptors

func (icf *interceptorsContainerFactory) generateMetachainHeaderInterceptor() ([]string, []process.Interceptor, error) {
	identifierHdr := factory.MetachainBlocksTopic
	metachainHeaderStorer := icf.store.GetStorer(dataRetriever.MetaBlockUnit)

	interceptor, err := interceptors.NewMetachainHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.MetaBlocks(),
		icf.dataPool.MetaHeadersNonces(),
		icf.tpsBenchmark,
		metachainHeaderStorer,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.chronologyValidator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}
