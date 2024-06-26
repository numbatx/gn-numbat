package block

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
	"github.com/numbatx/gn-numbat/core/partitioning"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/crypto/signing"
	"github.com/numbatx/gn-numbat/crypto/signing/kyber"
	"github.com/numbatx/gn-numbat/crypto/signing/kyber/singlesig"
	"github.com/numbatx/gn-numbat/data"
	dataBlock "github.com/numbatx/gn-numbat/data/block"
	"github.com/numbatx/gn-numbat/data/blockchain"
	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/data/state/addressConverters"
	"github.com/numbatx/gn-numbat/data/trie"
	"github.com/numbatx/gn-numbat/data/typeConverters/uint64ByteSlice"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/dataRetriever/dataPool"
	"github.com/numbatx/gn-numbat/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/numbatx/gn-numbat/dataRetriever/factory/metachain"
	factoryDataRetriever "github.com/numbatx/gn-numbat/dataRetriever/factory/shard"
	"github.com/numbatx/gn-numbat/dataRetriever/resolvers"
	"github.com/numbatx/gn-numbat/dataRetriever/shardedData"
	"github.com/numbatx/gn-numbat/display"
	"github.com/numbatx/gn-numbat/hashing/sha256"
	"github.com/numbatx/gn-numbat/integrationTests/mock"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/node"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/p2p/libp2p"
	"github.com/numbatx/gn-numbat/p2p/libp2p/discovery"
	"github.com/numbatx/gn-numbat/p2p/loadBalancer"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/block"
	"github.com/numbatx/gn-numbat/process/factory"
	metaProcess "github.com/numbatx/gn-numbat/process/factory/metachain"
	"github.com/numbatx/gn-numbat/process/factory/shard"
	"github.com/numbatx/gn-numbat/process/transaction"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/numbatx/gn-numbat/storage"
	"github.com/numbatx/gn-numbat/storage/memorydb"
)

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}
var testAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var testMultiSig = mock.NewMultiSigner(1)
var rootHash = []byte("root hash")

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node             *node.Node
	messenger        p2p.Messenger
	shardId          uint32
	accntState       state.AccountsAdapter
	blkc             data.ChainHandler
	blkProcessor     process.BlockProcessor
	sk               crypto.PrivateKey
	pk               crypto.PublicKey
	dPool            dataRetriever.PoolsHolder
	resFinder        dataRetriever.ResolversFinder
	headersRecv      int32
	miniblocksRecv   int32
	mutHeaders       sync.Mutex
	headersHashes    [][]byte
	headers          []data.HeaderHandler
	mutMiniblocks    sync.Mutex
	miniblocksHashes [][]byte
	miniblocks       []*dataBlock.MiniBlock
	metachainHdrRecv int32
	txsRecv          int32
}

func createTestShardChain() *blockchain.BlockChain {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createTestShardStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func createTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaHdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	metaHdrNonces, _ := dataPool.NewNonceToHashCacher(metaHdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 10, Type: storage.LRUCache}

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		metaHdrNonces,
	)

	return dPool
}

func createAccountsDB() *state.AccountsDB {
	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	targetShardId uint32,
	initialAddr string,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	dataRetriever.ResolversFinder,
	process.BlockProcessor,
	data.ChainHandler) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if shardCoordinator.ComputeId(addr) == targetShardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	blkc := createTestShardChain()
	store := createTestShardStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dataPacker, _ := partitioning.NewSizeDataPacker(testMarshalizer)

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		messenger,
		store,
		testMarshalizer,
		testHasher,
		keyGen,
		singleSigner,
		testMultiSig,
		dPool,
		testAddressConverter,
		&mock.ChronologyValidatorMock{},
		nil,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
		dataPacker,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolversFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)
	txProcessor, _ := transaction.NewTxProcessor(
		accntAdapter,
		testHasher,
		testAddressConverter,
		testMarshalizer,
		shardCoordinator,
	)

	blockProcessor, _ := block.NewShardProcessor(
		dPool,
		store,
		testHasher,
		testMarshalizer,
		txProcessor,
		accntAdapter,
		shardCoordinator,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{
			AddBlockCalled: func(headerHandler data.HeaderHandler) {
			},
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
		},
		func(destShardID uint32, txHashes [][]byte) {
			resolver, err := resolversFinder.CrossShardResolver(factory.TransactionTopic, destShardID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			err = resolver.(*resolvers.TxResolver).RequestDataFromHashArray(txHashes)
			if err != nil {
				fmt.Println(err.Error())
			}
		},
		func(shardId uint32, mbHash []byte) {
			resolver, err := resolversFinder.CrossShardResolver(factory.MiniBlocksTopic, shardId)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			err = resolver.RequestDataFromHash(mbHash)
			if err != nil {
				fmt.Println(err.Error())
			}
		},
	)

	n, err := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(singleSigner),
		node.WithTxSignPrivKey(sk),
		node.WithTxSignPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, sk, resolversFinder, blockProcessor, blkc
}

func createMessengerWithKadDht(ctx context.Context, initialAddr string) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessengerOnFreePort(
		ctx,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewKadDhtPeerDiscoverer(time.Second, "test", []string{initialAddr}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}
		return addr
	}
	return ""
}

func makeDisplayTable(nodes []*testNode) string {
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))
	for idx, n := range nodes {
		buffPk, _ := n.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%d", n.shardId),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.txsRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.miniblocksRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.headersRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.metachainHdrRecv)),
				fmt.Sprintf("%d / %d", len(n.messenger.ConnectedPeersOnTopic(factory.TransactionTopic+"_"+
					fmt.Sprintf("%d", n.shardId))), len(n.messenger.ConnectedPeers())),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

func displayAndStartNodes(nodes []*testNode) {
	for _, n := range nodes {
		skBuff, _ := n.sk.ToByteArray()
		pkBuff, _ := n.pk.ToByteArray()

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			n.shardId,
			hex.EncodeToString(skBuff),
			hex.EncodeToString(pkBuff),
		)
		_ = n.node.Start()
		_ = n.node.P2PBootstrap()
	}
}

func createNodes(
	numOfShards int,
	nodesPerShard int,
	serviceID string,
) []*testNode {

	//first node generated will have is pk belonging to firstSkShardId
	numMetaChainNodes := 1
	nodes := make([]*testNode, int(numOfShards)*nodesPerShard+numMetaChainNodes)

	idx := 0
	for shardId := 0; shardId < numOfShards; shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			testNode := &testNode{
				dPool:   createTestShardDataPool(),
				shardId: uint32(shardId),
			}

			shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
			accntAdapter := createAccountsDB()
			n, mes, sk, resFinder, blkProcessor, blkc := createNetNode(
				testNode.dPool,
				accntAdapter,
				shardCoordinator,
				testNode.shardId,
				serviceID,
			)
			_ = n.CreateShardedStores()

			testNode.node = n
			testNode.sk = sk
			testNode.messenger = mes
			testNode.pk = sk.GeneratePublic()
			testNode.resFinder = resFinder
			testNode.accntState = accntAdapter
			testNode.blkProcessor = blkProcessor
			testNode.blkc = blkc
			testNode.dPool.Headers().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.headersRecv, 1)
				testNode.mutHeaders.Lock()
				testNode.headersHashes = append(testNode.headersHashes, key)
				header, _ := testNode.dPool.Headers().Peek(key)
				testNode.headers = append(testNode.headers, header.(data.HeaderHandler))
				testNode.mutHeaders.Unlock()
			})
			testNode.dPool.MiniBlocks().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.miniblocksRecv, 1)
				testNode.mutMiniblocks.Lock()
				testNode.miniblocksHashes = append(testNode.miniblocksHashes, key)
				miniblock, _ := testNode.dPool.MiniBlocks().Peek(key)
				testNode.miniblocks = append(testNode.miniblocks, miniblock.(*dataBlock.MiniBlock))
				testNode.mutMiniblocks.Unlock()
			})
			testNode.dPool.MetaBlocks().RegisterHandler(func(key []byte) {
				fmt.Printf("Got metachain header: %v\n", base64.StdEncoding.EncodeToString(key))
				atomic.AddInt32(&testNode.metachainHdrRecv, 1)
			})
			testNode.dPool.Transactions().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.txsRecv, 1)
			})

			nodes[idx] = testNode
			idx++
		}
	}

	shardCoordinatorMeta, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), sharding.MetachainShardId)
	tn := createMetaNetNode(
		createTestMetaDataPool(),
		createAccountsDB(),
		shardCoordinatorMeta,
		serviceID,
	)
	for i := 0; i < numMetaChainNodes; i++ {
		idx := i + int(numOfShards)*nodesPerShard
		nodes[idx] = tn
	}

	return nodes
}

func getMiniBlocksHashesFromShardIds(body dataBlock.Body, shardIds ...uint32) [][]byte {
	hashes := make([][]byte, 0)

	for _, miniblock := range body {
		for _, shardId := range shardIds {
			if miniblock.ReceiverShardID == shardId {
				buff, _ := testMarshalizer.Marshal(miniblock)
				hashes = append(hashes, testHasher.Compute(string(buff)))
			}
		}
	}

	return hashes
}

func equalSlices(slice1 [][]byte, slice2 [][]byte) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	//check slice1 has all elements in slice2
	for _, buff1 := range slice1 {
		found := false
		for _, buff2 := range slice2 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	//check slice2 has all elements in slice1
	for _, buff2 := range slice2 {
		found := false
		for _, buff1 := range slice1 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func uint32InSlice(searched uint32, list []uint32) bool {
	for _, val := range list {
		if val == searched {
			return true
		}
	}
	return false
}

func generatePrivateKeyInShardId(
	coordinator sharding.Coordinator,
	shardId uint32,
) crypto.PrivateKey {

	suite := kyber.NewBlakeSHA256Ed25519()
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		buff, _ := pk.ToByteArray()
		addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(buff)

		if coordinator.ComputeId(addr) == shardId {
			return sk
		}

		sk, pk = keyGen.GeneratePair()
	}
}

func createTestMetaChain() data.ChainHandler {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	metaChain, _ := blockchain.NewMetaChain(
		badBlockCache,
	)
	metaChain.GenesisBlock = &dataBlock.MetaBlock{}

	return metaChain
}

func createTestMetaStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaPeerDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaShardDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func createTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 10000, Type: storage.LRUCache}
	miniblockHashes, _ := shardedData.NewShardedData(cacherCfg)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	shardHeaders, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaBlockNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	metaBlockNonces, _ := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		miniblockHashes,
		shardHeaders,
		metaBlockNonces,
	)

	return dPool
}

func createMetaNetNode(
	dPool dataRetriever.MetaPoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	initialAddr string,
) *testNode {

	tn := testNode{}

	tn.messenger = createMessengerWithKadDht(context.Background(), initialAddr)
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	tn.blkc = createTestMetaChain()
	store := createTestMetaStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		testHasher,
		testMultiSig,
		dPool,
		&mock.ChronologyValidatorMock{},
		nil,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolvers, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	blkProc, _ := block.NewMetaProcessor(
		accntAdapter,
		dPool,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		shardCoordinator,
		testHasher,
		testMarshalizer,
		store,
		func(shardId uint32, hdrHash []byte) {},
	)
	_ = blkProc.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator))
	tn.blkProcessor = blkProc

	n, err := node.NewNode(
		node.WithMessenger(tn.messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithMetaDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(tn.blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(singleSigner),
		node.WithPrivKey(sk),
		node.WithPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolvers),
		node.WithBlockProcessor(tn.blkProcessor),
		node.WithDataStore(store),
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	tn.node = n
	tn.sk = sk
	tn.pk = pk
	tn.accntState = accntAdapter
	tn.shardId = sharding.MetachainShardId

	dPool.MetaChainBlocks().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.metachainHdrRecv, 1)
	})
	dPool.ShardHeaders().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.headersRecv, 1)
		tn.mutHeaders.Lock()
		metaHeader, _ := dPool.ShardHeaders().Peek(key)
		tn.headers = append(tn.headers, metaHeader.(data.HeaderHandler))
		tn.mutHeaders.Unlock()
	})

	return &tn
}

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = createGenesisBlock(shardId)
	}

	return genesisBlocks
}

func createGenesisBlock(shardId uint32) *dataBlock.Header {
	return &dataBlock.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       shardId,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}
