package spos

import (
	"github.com/numbatx/gn-numbat/consensus"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/ntp"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/sharding"
)

// ConsensusCoreHandler encapsulates all needed Data for the Consensus
type ConsensusCoreHandler interface {
	// Blockchain gets the ChainHandler stored in the ConsensusCore
	Blockchain() data.ChainHandler
	// BlockProcessor gets the BlockProcessor stored in the ConsensusCore
	BlockProcessor() process.BlockProcessor
	// BootStrapper gets the Bootstrapper stored in the ConsensusCore
	BootStrapper() process.Bootstrapper
	// Chronology gets the ChronologyHandler stored in the ConsensusCore
	Chronology() consensus.ChronologyHandler
	// Hasher gets the Hasher stored in the ConsensusCore
	Hasher() hashing.Hasher
	// Marshalizer gets the Marshalizer stored in the ConsensusCore
	Marshalizer() marshal.Marshalizer
	// MultiSigner gets the MultiSigner stored in the ConsensusCore
	MultiSigner() crypto.MultiSigner
	// Rounder gets the Rounder stored in the ConsensusCore
	Rounder() consensus.Rounder
	// ShardCoordinator gets the Coordinator stored in the ConsensusCore
	ShardCoordinator() sharding.Coordinator
	// SyncTimer gets the SyncTimer stored in the ConsensusCore
	SyncTimer() ntp.SyncTimer
	// ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusCore
	ValidatorGroupSelector() consensus.ValidatorGroupSelector
	// RandomnessPrivateKey returns the private key stored in the ConsensusStore used for randomness generation
	RandomnessPrivateKey() crypto.PrivateKey
	// RandomnessSingleSigner returns the single signer stored in the ConsensusStore used for randomness generation
	RandomnessSingleSigner() crypto.SingleSigner
}

// ConsensusService encapsulates the methods specifically for a consensus type (bls, bn)
// and will be used in the sposWorker
type ConsensusService interface {
	//InitReceivedMessages initializes the MessagesType map for all messages for the current ConsensusService
	InitReceivedMessages() map[consensus.MessageType][]*consensus.Message
	//GetStringValue gets the name of the messageType
	GetStringValue(consensus.MessageType) string
	//GetSubroundName gets the subround name for the subround id provided
	GetSubroundName(int) string
	//GetMessageRange provides the MessageType range used in checks by the consensus
	GetMessageRange() []consensus.MessageType
	//CanProceed returns if the current messageType can proceed further if previous subrounds finished
	CanProceed(*ConsensusState, consensus.MessageType) bool
	//IsMessageWithBlockHeader returns if the current messageType is about block header
	IsMessageWithBlockHeader(consensus.MessageType) bool
}

// SubroundsFactory encapsulates the methods specifically for a subrounds factory type (bls, bn)
// for different consensus types
type SubroundsFactory interface {
	GenerateSubrounds() error
}

// WorkerHandler represents the interface for the SposWorker
type WorkerHandler interface {
	//AddReceivedMessageCall adds a new handler function for a received messege type
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool)
	//RemoveAllReceivedMessagesCalls removes all the functions handlers
	RemoveAllReceivedMessagesCalls()
	//ProcessReceivedMessage method redirects the received message to the channel which should handle it
	ProcessReceivedMessage(message p2p.MessageP2P) error
	//SendConsensusMessage sends the consensus message
	SendConsensusMessage(cnsDta *consensus.Message) bool
	//Extend does an extension for the subround with subroundId
	Extend(subroundId int)
	//GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
	GetConsensusStateChangedChannel() chan bool
	//BroadcastBlock does a broadcast of the blockBody and blockHeader
	BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error
	//ExecuteStoredMessages tries to execute all the messages received which are valid for execution
	ExecuteStoredMessages()
	//BroadcastUnnotarisedBlocks broadcasts all blocks which are not notarised yet
	BroadcastUnnotarisedBlocks()
}
