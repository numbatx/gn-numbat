package block

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/numbatx/gn-numbat/data/block"
	"github.com/numbatx/gn-numbat/dataRetriever/resolvers"
	"github.com/numbatx/gn-numbat/hashing/sha256"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process/factory"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptHeaderByNonceWithMemMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requestor:")
	nRequestor, mesRequestor, _, resolversFinder := createNetNode(
		dPoolRequestor,
		createAccountsDB(),
		shardCoordinator)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := createNetNode(
		dPoolResolver,
		createAccountsDB(),
		shardCoordinator)

	nRequestor.Start()
	nResolver.Start()
	defer func() {
		_ = nRequestor.Stop()
		_ = nResolver.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := mesRequestor.ConnectToPeer(getConnectableAddress(mesResolver))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate a header
	hdr := block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardId:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
	}

	hdrBuff, _ := marshalizer.Marshal(&hdr)
	hdrHash := hasher.Compute(string(hdrBuff))

	//Step 2. resolver has the header
	dPoolResolver.Headers().HasOrAdd(hdrHash, &hdr)
	dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash)

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequestor.Headers().Peek(key)

		if reflect.DeepEqual(hdrStored, &hdr) && hdr.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, hdrStored, &hdr)

	})

	//Step 4. request header
	res, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	hdrResolver.RequestDataFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
