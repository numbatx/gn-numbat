package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/numbatx/gn-numbat/crypto/signing/kyber/singlesig"
	"github.com/numbatx/gn-numbat/data/state"

	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/data/block"
	"github.com/numbatx/gn-numbat/data/transaction"
	"github.com/numbatx/gn-numbat/node"
	"github.com/numbatx/gn-numbat/sharding"
	"github.com/stretchr/testify/assert"
)

func TestShouldProcessBlocksInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Step 1. Setup nodes...")
	numOfShards := 6
	nodesPerShard := 3

	senderShard := uint32(0)
	recvShards := []uint32{1, 2}

	valMinting := big.NewInt(100)
	valToTransferPerTx := big.NewInt(2)

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodes(
		numOfShards,
		nodesPerShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println("Step 2. Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	txToGenerateInEachMiniBlock := 3

	proposerNode := nodes[0]

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, 3)
	receiversPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i] = generatePrivateKeyInShardId(generateCoordinator, senderShard)

		//receivers in same shard with the sender
		sk := generatePrivateKeyInShardId(generateCoordinator, senderShard)
		receiversPrivateKeys[senderShard] = append(receiversPrivateKeys[senderShard], sk)
		//receivers in other shards
		for _, shardId := range recvShards {
			sk = generatePrivateKeyInShardId(generateCoordinator, shardId)
			receiversPrivateKeys[shardId] = append(receiversPrivateKeys[shardId], sk)
		}
	}

	fmt.Println("Step 3. Generating transactions...")
	generateAndDisseminateTxs(proposerNode.node, sendersPrivateKeys, receiversPrivateKeys, valToTransferPerTx)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	fmt.Println("Step 4. Minting sender addresses...")
	createMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	fmt.Println("Step 5. Proposer creates block body and header with all available transactions...")
	blockBody, blockHeader := proposeBlock(t, proposerNode)

	fmt.Println("Step 6. Proposer disseminates header, block body and miniblocks...")
	proposerNode.node.BroadcastShardBlock(blockBody, blockHeader)
	proposerNode.node.BroadcastShardHeader(blockHeader)
	fmt.Println("Delaying for disseminating miniblocks and header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 7. NodesSetup from proposer's shard will have to successfully process the block sent by the proposer...")
	fmt.Println(makeDisplayTable(nodes))
	for _, n := range nodes {
		isNodeInSenderShardAndNotProposer := n.shardId == senderShard && n != proposerNode
		if isNodeInSenderShardAndNotProposer {
			n.blkc.SetGenesisHeaderHash(n.headers[0].GetPrevHash())
			err := n.blkProcessor.ProcessBlock(
				n.blkc,
				n.headers[0],
				block.Body(n.miniblocks),
				func() time.Duration {
					//fair enough to process a few transactions
					return time.Second * 2
				},
			)

			assert.Nil(t, err)
		}
	}

	fmt.Println("Step 7. Metachain processes the received header...")
	metaNode := nodes[len(nodes)-1]
	_, metaHeader := proposeMetaBlock(t, metaNode)
	metaNode.node.BroadcastMetaBlock(nil, metaHeader)
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 8. Test nodes from proposer shard to have the correct balances...")
	for _, n := range nodes {
		isNodeInSenderShard := n.shardId == senderShard
		if !isNodeInSenderShard {
			continue
		}

		//test sender balances
		for _, sk := range sendersPrivateKeys {
			valTransferred := big.NewInt(0).Mul(valToTransferPerTx, big.NewInt(int64(len(receiversPrivateKeys))))
			valRemaining := big.NewInt(0).Sub(valMinting, valTransferred)
			testPrivateKeyHasBalance(t, n, sk, valRemaining)
		}
		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[proposerNode.shardId] {
			testPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}

	fmt.Println("Step 9. First nodes from receiver shards assemble header/body blocks and broadcast them...")
	firstReceiverNodes := make([]*testNode, 0)
	//get first nodes from receiver shards
	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]
		firstReceiverNodes = append(firstReceiverNodes, receiverProposer)

		body, header := proposeBlock(t, receiverProposer)
		receiverProposer.node.BroadcastShardBlock(body, header)
		receiverProposer.node.BroadcastShardHeader(header)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 10. NodesSetup from receivers shards will have to successfully process the block sent by their proposer...")
	fmt.Println(makeDisplayTable(nodes))
	for _, n := range nodes {
		if n.shardId == sharding.MetachainShardId {
			continue
		}

		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.shardId == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		for _, proposerReceiver := range firstReceiverNodes {
			if proposerReceiver == n {
				isNodeInReceiverShardAndNotProposer = false
			}
		}

		if isNodeInReceiverShardAndNotProposer {
			if len(n.headers) > 0 {
				n.blkc.SetGenesisHeaderHash(n.headers[0].GetPrevHash())
				err := n.blkProcessor.ProcessBlock(
					n.blkc,
					n.headers[0],
					block.Body(n.miniblocks),
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.blkProcessor.CommitBlock(n.blkc, n.headers[0], block.Body(n.miniblocks))
			}
		}
	}

	fmt.Println("Step 11. Test nodes from receiver shards to have the correct balances...")
	for _, n := range nodes {
		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.shardId == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		if !isNodeInReceiverShardAndNotProposer {
			continue
		}

		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[n.shardId] {
			testPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}

}

func generateAndDisseminateTxs(
	n *node.Node,
	senders []crypto.PrivateKey,
	receiversPrivateKeys map[uint32][]crypto.PrivateKey,
	valToTransfer *big.Int,
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := uint64(0)
		for _, recvPrivateKeys := range receiversPrivateKeys {
			receiverKey := recvPrivateKeys[i]
			tx := generateTransferTx(incrementalNonce, senderKey, receiverKey, valToTransfer)
			n.SendTransaction(
				tx.Nonce,
				hex.EncodeToString(tx.SndAddr),
				hex.EncodeToString(tx.RcvAddr),
				tx.Value,
				string(tx.Data),
				tx.Signature,
			)
			incrementalNonce++
		}
	}
}

func generateTransferTx(
	nonce uint64,
	sender crypto.PrivateKey,
	receiver crypto.PrivateKey,
	valToTransfer *big.Int) *transaction.Transaction {

	tx := transaction.Transaction{
		Nonce:   nonce,
		Value:   valToTransfer,
		RcvAddr: skToPk(receiver),
		SndAddr: skToPk(sender),
		Data:    make([]byte, 0),
	}
	txBuff, _ := testMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}
	tx.Signature, _ = signer.Sign(sender, txBuff)

	return &tx
}

func skToPk(sk crypto.PrivateKey) []byte {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	return pkBuff
}

func createMintingForSenders(
	nodes []*testNode,
	senderShard uint32,
	sendersPrivateKeys []crypto.PrivateKey,
	value *big.Int,
) {

	for _, n := range nodes {
		//only sender shard nodes will be minted
		if n.shardId != senderShard {
			continue
		}

		for _, sk := range sendersPrivateKeys {
			pkBuff, _ := sk.GeneratePublic().ToByteArray()
			adr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
			account, _ := n.accntState.GetAccountWithJournal(adr)
			account.(*state.Account).SetBalanceWithJournal(value)
		}

		n.accntState.Commit()
	}
}

func testPrivateKeyHasBalance(t *testing.T, n *testNode, sk crypto.PrivateKey, expectedBalance *big.Int) {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
	account, _ := n.accntState.GetExistingAccount(addr)
	assert.Equal(t, expectedBalance, account.(*state.Account).Balance)
}

func proposeBlock(t *testing.T, proposer *testNode) (data.BodyHandler, data.HeaderHandler) {
	blockBody, err := proposer.blkProcessor.CreateBlockBody(1, func() bool {
		return true
	})
	assert.Nil(t, err)

	blockHeader, err := proposer.blkProcessor.CreateBlockHeader(blockBody, 1, func() bool {
		return true
	})
	assert.Nil(t, err)

	blockHeader.SetNonce(1)
	blockHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := testMultiSig.AggregateSigs(nil)
	blockHeader.SetSignature(sig)
	buffGenesis, _ := testMarshalizer.Marshal(createGenesisBlock(proposer.shardId))
	blockHeader.SetPrevHash(testHasher.Compute(string(buffGenesis)))
	blockHeader.SetPrevRandSeed(rootHash)
	blockHeader.SetRandSeed(sig)
	blockHeader.SetRound(1)

	return blockBody, blockHeader
}

func proposeMetaBlock(t *testing.T, proposer *testNode) (data.BodyHandler, data.HeaderHandler) {
	metaHeader, err := proposer.blkProcessor.CreateBlockHeader(nil, 2, func() bool {
		return true
	})
	assert.Nil(t, err)

	metaHeader.SetNonce(1)
	metaHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := testMultiSig.AggregateSigs(nil)
	metaHeader.SetSignature(sig)
	buffGenesis, _ := testMarshalizer.Marshal(proposer.blkc.GetGenesisHeader())
	metaHeader.SetPrevHash(testHasher.Compute(string(buffGenesis)))
	metaHeader.SetRandSeed(sig)

	return nil, metaHeader
}
