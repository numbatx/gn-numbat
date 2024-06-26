package spos

import (
	"github.com/numbatx/gn-numbat/consensus"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/ntp"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/sharding"
)

// ConsensusCore implements ConsensusCoreHandler and provides access to common functionalities
//
//	for the rest of the consensus structures
type ConsensusCore struct {
	blockChain             data.ChainHandler
	blockProcessor         process.BlockProcessor
	bootstraper            process.Bootstrapper
	chronologyHandler      consensus.ChronologyHandler
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	blsPrivateKey          crypto.PrivateKey
	blsSingleSigner        crypto.SingleSigner
	multiSigner            crypto.MultiSigner
	rounder                consensus.Rounder
	shardCoordinator       sharding.Coordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector consensus.ValidatorGroupSelector
}

// NewConsensusCore creates a new ConsensusCore instance
func NewConsensusCore(
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	chronologyHandler consensus.ChronologyHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	blsPrivateKey crypto.PrivateKey,
	blsSingleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector) (*ConsensusCore, error) {

	consensusCore := &ConsensusCore{
		blockChain,
		blockProcessor,
		bootstraper,
		chronologyHandler,
		hasher,
		marshalizer,
		blsPrivateKey,
		blsSingleSigner,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		validatorGroupSelector,
	}

	err := ValidateConsensusCore(consensusCore)

	if err != nil {
		return nil, err
	}
	return consensusCore, nil
}

// Blockchain gets the ChainHandler stored in the ConsensusCore
func (cc *ConsensusCore) Blockchain() data.ChainHandler {
	return cc.blockChain
}

// BlockProcessor gets the BlockProcessor stored in the ConsensusCore
func (cc *ConsensusCore) BlockProcessor() process.BlockProcessor {
	return cc.blockProcessor
}

// BootStrapper gets the Bootstrapper stored in the ConsensusCore
func (cc *ConsensusCore) BootStrapper() process.Bootstrapper {
	return cc.bootstraper
}

// Chronology gets the ChronologyHandler stored in the ConsensusCore
func (cc *ConsensusCore) Chronology() consensus.ChronologyHandler {
	return cc.chronologyHandler
}

// Hasher gets the Hasher stored in the ConsensusCore
func (cc *ConsensusCore) Hasher() hashing.Hasher {
	return cc.hasher
}

// Marshalizer gets the Marshalizer stored in the ConsensusCore
func (cc *ConsensusCore) Marshalizer() marshal.Marshalizer {
	return cc.marshalizer
}

// MultiSigner gets the MultiSigner stored in the ConsensusCore
func (cc *ConsensusCore) MultiSigner() crypto.MultiSigner {
	return cc.multiSigner
}

// Rounder gets the Rounder stored in the ConsensusCore
func (cc *ConsensusCore) Rounder() consensus.Rounder {
	return cc.rounder
}

// ShardCoordinator gets the Coordinator stored in the ConsensusCore
func (cc *ConsensusCore) ShardCoordinator() sharding.Coordinator {
	return cc.shardCoordinator
}

// SyncTimer gets the SyncTimer stored in the ConsensusCore
func (cc *ConsensusCore) SyncTimer() ntp.SyncTimer {
	return cc.syncTimer
}

// ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusCore
func (cc *ConsensusCore) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cc.validatorGroupSelector
}

// RandomnessPrivateKey returns the BLS private key stored in the ConsensusStore
func (cc *ConsensusCore) RandomnessPrivateKey() crypto.PrivateKey {
	return cc.blsPrivateKey
}

// RandomnessSingleSigner returns the bls single signer stored in the ConsensusStore
func (cc *ConsensusCore) RandomnessSingleSigner() crypto.SingleSigner {
	return cc.blsSingleSigner
}
