package transaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/numbatx/gn-numbat/crypto/signing/kyber/singlesig"
	"github.com/numbatx/gn-numbat/data/transaction"
	"github.com/numbatx/gn-numbat/hashing/sha256"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/process/factory"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTransactionWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requestor:")
	nRequestor, mesRequestor, sk1, resolversFinder := createNetNode(
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

	buffPk1, _ := sk1.GeneratePublic().ToByteArray()

	//Step 1. Generate a signed transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: buffPk1,
		Data:    []byte("tx notarized data"),
	}

	txBuff, _ := marshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}

	tx.Signature, _ = signer.Sign(sk1, txBuff)

	signedTxBuff, _ := marshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(signedTxBuff))

	chanDone := make(chan bool)

	txHash := hasher.Compute(string(signedTxBuff))

	//step 2. wire up a received handler for requestor
	dPoolRequestor.Transactions().RegisterHandler(func(key []byte) {
		txStored, _ := dPoolRequestor.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	dPoolResolver.Transactions().AddData(
		txHash,
		&tx,
		process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
	)

	//Step 4. request tx
	txResolver, _ := resolversFinder.IntraShardResolver(factory.TransactionTopic)
	err = txResolver.RequestDataFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
