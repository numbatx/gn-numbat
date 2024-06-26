package facade

import (
	"math/big"

	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/data/transaction"
	"github.com/numbatx/gn-numbat/node/external"
	"github.com/numbatx/gn-numbat/node/heartbeat"
)

// NodeWrapper contains all functions that a node should contain.
type NodeWrapper interface {

	// Start will create a new messenger and and set up the Node state as running
	Start() error

	// Stop closes the messenger and undos everything done in Start
	Stop() error

	// P2PBootstrap starts the peer discovery process and peer connection filtering
	P2PBootstrap() error

	//IsRunning returns if the underlying node is running
	IsRunning() bool

	// StartConsensus will start the consesus service for the current node
	StartConsensus() error

	//GetBalance returns the balance for a specific address
	GetBalance(address string) (*big.Int, error)

	//GenerateTransaction generates a new transaction with sender, receiver, amount and code
	GenerateTransaction(senderHex string, receiverHex string, amount *big.Int, code string) (*transaction.Transaction, error)

	//SendTransaction will send a new transaction on the topic channel
	SendTransaction(nonce uint64, senderHex string, receiverHex string, value *big.Int, transactionData string, signature []byte) (*transaction.Transaction, error)

	//GetTransaction gets the transaction
	GetTransaction(hash string) (*transaction.Transaction, error)

	// GetCurrentPublicKey gets the current nodes public Key
	GetCurrentPublicKey() string

	// GenerateAndSendBulkTransactions generates a number of nrTransactions of amount value
	//  for the receiver destination
	GenerateAndSendBulkTransactions(string, *big.Int, uint64) error

	// GenerateAndSendBulkTransactionsOneByOne generates a number of nrTransactions of amount value
	//  for the receiver destination in a one-by-one fashion
	GenerateAndSendBulkTransactionsOneByOne(string, *big.Int, uint64) error

	// GetAccount returns an accountResponse containing information
	//  about the account corelated with provided address
	GetAccount(address string) (*state.Account, error)

	// GetHeartbeats returns the heartbeat status for each public key defined in genesis.json
	GetHeartbeats() []heartbeat.PubKeyHeartbeat
}

// ExternalResolver defines what functionality can be exposed to an external component (REST API, RPC, etc.)
type ExternalResolver interface {
	RecentNotarizedBlocks(maxShardHeadersNum int) ([]*external.BlockHeader, error)
	RetrieveShardBlock(blockHash []byte) (*external.ShardBlockInfo, error)
}
