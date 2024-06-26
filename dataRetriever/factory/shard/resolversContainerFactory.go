package shard

import (
	"github.com/numbatx/gn-numbat/core/random"
	"github.com/numbatx/gn-numbat/data/typeConverters"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/dataRetriever/factory/containers"
	"github.com/numbatx/gn-numbat/dataRetriever/resolvers"
	"github.com/numbatx/gn-numbat/dataRetriever/resolvers/topicResolverSender"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process/factory"
	"github.com/numbatx/gn-numbat/sharding"
)

type resolversContainerFactory struct {
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	dataPools                dataRetriever.PoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	intRandomizer            dataRetriever.IntRandomizer
	dataPacker               dataRetriever.DataPacker
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.PoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	dataPacker dataRetriever.DataPacker,
) (*resolversContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, dataRetriever.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, dataRetriever.ErrNilMessenger
	}
	if store == nil {
		return nil, dataRetriever.ErrNilTxStorage
	}
	if marshalizer == nil {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if dataPools == nil {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if uint64ByteSliceConverter == nil {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if dataPacker == nil {
		return nil, dataRetriever.ErrNilDataPacker
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               dataPacker,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, resolverSlice, err := rcf.generateTxResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateHdrResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generatePeerChBlockBodyResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateMetachainShardHeaderResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateMetablockHeaderResolver()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver dataRetriever.Resolver,
	createChannel bool,
) (dataRetriever.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

//------- Tx resolvers

func (rcf *resolversContainerFactory) generateTxResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneTxResolver(identifierTx)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createOneTxResolver(identifier string) (dataRetriever.Resolver, error) {
	txStorer := rcf.store.GetStorer(dataRetriever.TransactionUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewTxResolver(
		resolverSender,
		rcf.dataPools.Transactions(),
		txStorer,
		rcf.marshalizer,
		rcf.dataPacker,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Hdr resolver

func (rcf *resolversContainerFactory) generateHdrResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	//only one intrashard header topic
	identifierHdr := factory.HeadersTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	hdrStorer := rcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifierHdr,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, nil, err
	}
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.Headers(),
		rcf.dataPools.HeadersNonces(),
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}
	//add on the request topic
	_, err = rcf.createTopicAndAssignHandler(
		identifierHdr+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	err = rcf.createTopicHeadersForMetachain()
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []dataRetriever.Resolver{resolver}, nil
}

func (rcf *resolversContainerFactory) createTopicHeadersForMetachain() error {
	shardC := rcf.shardCoordinator
	identifierHdr := factory.ShardHeadersForMetachainTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	return rcf.messenger.CreateTopic(identifierHdr, true)
}

//------- MiniBlocks resolvers

func (rcf *resolversContainerFactory) generateMiniBlocksResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneMiniBlocksResolver(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createOneMiniBlocksResolver(identifier string) (dataRetriever.Resolver, error) {
	miniBlocksStorer := rcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPools.MiniBlocks(),
		miniBlocksStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		txBlkResolver,
		false)
}

//------- PeerChBlocks resolvers

func (rcf *resolversContainerFactory) generatePeerChBlockBodyResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := factory.PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	peerBlockBodyStorer := rcf.store.GetStorer(dataRetriever.PeerChangesUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifierPeerCh,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, nil, err
	}

	resolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPools.MiniBlocks(),
		peerBlockBodyStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, nil, err
	}
	//add on the request topic
	_, err = rcf.createTopicAndAssignHandler(
		identifierPeerCh+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []dataRetriever.Resolver{resolver}, nil
}

//------- MetachainShardHeaderResolvers

func (rcf *resolversContainerFactory) generateMetachainShardHeaderResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	//only one metachain header topic
	//example: shardHeadersForMetachain_0
	identifierHdr := factory.ShardHeadersForMetachainTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	hdrStorer := rcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifierHdr,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, nil, err
	}

	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.Headers(),
		rcf.dataPools.HeadersNonces(),
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}

	//add on the request topic
	_, err = rcf.createTopicAndAssignHandler(
		identifierHdr+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []dataRetriever.Resolver{resolver}, nil
}

//------- MetaBlockHeaderResolvers

func (rcf *resolversContainerFactory) generateMetablockHeaderResolver() ([]string, []dataRetriever.Resolver, error) {
	//only one metachain header block topic
	//this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer := rcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifierHdr,
		rcf.marshalizer,
		rcf.intRandomizer,
	)
	if err != nil {
		return nil, nil, err
	}

	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.MetaBlocks(),
		rcf.dataPools.MetaHeadersNonces(),
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}

	//add on the request topic
	_, err = rcf.createTopicAndAssignHandler(
		identifierHdr+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []dataRetriever.Resolver{resolver}, nil
}
