package facade

import (
	"math/big"
	"sync"

	"github.com/numbatx/gn-numbat/api"
	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/core/statistics"
	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/data/transaction"
	"github.com/numbatx/gn-numbat/node/external"
	"github.com/numbatx/gn-numbat/node/heartbeat"
	"github.com/numbatx/gn-numbat/ntp"
)

// NumbatNodeFacade represents a facade for grouping the functionality for node, transaction and address
type NumbatNodeFacade struct {
	node         NodeWrapper
	resolver     ExternalResolver
	syncer       ntp.SyncTimer
	log          *logger.Logger
	tpsBenchmark *statistics.TpsBenchmark
}

// NewNumbatNodeFacade creates a new Facade with a NodeWrapper
func NewNumbatNodeFacade(node NodeWrapper, resolver ExternalResolver) *NumbatNodeFacade {
	if node == nil {
		return nil
	}
	if resolver == nil {
		return nil
	}

	return &NumbatNodeFacade{
		node:     node,
		resolver: resolver,
	}
}

// SetLogger sets the current logger
func (ef *NumbatNodeFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

// SetSyncer sets the current syncer
func (ef *NumbatNodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	ef.syncer = syncer
}

// SetTpsBenchmark sets the tps benchmark handler
func (ef *NumbatNodeFacade) SetTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark) {
	ef.tpsBenchmark = tpsBenchmark
}

// TpsBenchmark returns the tps benchmark handler
func (ef *NumbatNodeFacade) TpsBenchmark() *statistics.TpsBenchmark {
	return ef.tpsBenchmark
}

// StartNode starts the underlying node
func (ef *NumbatNodeFacade) StartNode() error {
	err := ef.node.Start()
	if err != nil {
		return err
	}

	err = ef.node.StartConsensus()
	return err
}

// StopNode stops the underlying node
func (ef *NumbatNodeFacade) StopNode() error {
	return ef.node.Stop()
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (ef *NumbatNodeFacade) StartBackgroundServices(wg *sync.WaitGroup) {
	wg.Add(1)
	go ef.startRest(wg)
}

// IsNodeRunning gets if the underlying node is running
func (ef *NumbatNodeFacade) IsNodeRunning() bool {
	return ef.node.IsRunning()
}

func (ef *NumbatNodeFacade) startRest(wg *sync.WaitGroup) {
	defer wg.Done()

	ef.log.Info("Starting web server...")
	err := api.Start(ef)
	if err != nil {
		ef.log.Error("Could not start webserver", err.Error())
	}
}

// GetBalance gets the current balance for a specified address
func (ef *NumbatNodeFacade) GetBalance(address string) (*big.Int, error) {
	return ef.node.GetBalance(address)
}

// GenerateTransaction generates a transaction from a sender, receiver, value and data
func (ef *NumbatNodeFacade) GenerateTransaction(senderHex string, receiverHex string, value *big.Int,
	data string) (*transaction.Transaction,
	error) {
	return ef.node.GenerateTransaction(senderHex, receiverHex, value, data)
}

// SendTransaction will send a new transaction on the topic channel
func (ef *NumbatNodeFacade) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value *big.Int,
	transactionData string,
	signature []byte,
) (*transaction.Transaction, error) {

	return ef.node.SendTransaction(nonce, senderHex, receiverHex, value, transactionData, signature)
}

// GetTransaction gets the transaction with a specified hash
func (ef *NumbatNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (ef *NumbatNodeFacade) GetAccount(address string) (*state.Account, error) {
	return ef.node.GetAccount(address)
}

// GetCurrentPublicKey gets the current nodes public Key
func (ef *NumbatNodeFacade) GetCurrentPublicKey() string {
	return ef.node.GetCurrentPublicKey()
}

// GenerateAndSendBulkTransactions generates a number of nrTransactions of amount value
// for the receiver destination
func (ef *NumbatNodeFacade) GenerateAndSendBulkTransactions(
	destination string,
	value *big.Int,
	nrTransactions uint64,
) error {

	return ef.node.GenerateAndSendBulkTransactions(destination, value, nrTransactions)
}

// GenerateAndSendBulkTransactionsOneByOne generates a number of nrTransactions of amount value
// for the receiver destination in a one by one fashion
func (ef *NumbatNodeFacade) GenerateAndSendBulkTransactionsOneByOne(
	destination string,
	value *big.Int,
	nrTransactions uint64,
) error {

	return ef.node.GenerateAndSendBulkTransactionsOneByOne(destination, value, nrTransactions)
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (ef *NumbatNodeFacade) GetHeartbeats() ([]heartbeat.PubKeyHeartbeat, error) {
	hbStatus := ef.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// RecentNotarizedBlocks computes last notarized [maxShardHeadersNum] shard headers (by metachain node)
func (ef *NumbatNodeFacade) RecentNotarizedBlocks(maxShardHeadersNum int) ([]*external.BlockHeader, error) {
	return ef.resolver.RecentNotarizedBlocks(maxShardHeadersNum)
}

// RetrieveShardBlock retrieves a shard block info containing header and transactions
func (ef *NumbatNodeFacade) RetrieveShardBlock(blockHash []byte) (*external.ShardBlockInfo, error) {
	return ef.resolver.RetrieveShardBlock(blockHash)
}
