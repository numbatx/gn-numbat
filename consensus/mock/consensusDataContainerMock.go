package mock

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

type ConsensusCoreMock struct {
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

func (cdc *ConsensusCoreMock) Blockchain() data.ChainHandler {
	return cdc.blockChain
}

func (cdc *ConsensusCoreMock) BlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}

func (cdc *ConsensusCoreMock) BootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}

func (cdc *ConsensusCoreMock) Chronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}

func (cdc *ConsensusCoreMock) Hasher() hashing.Hasher {
	return cdc.hasher
}

func (cdc *ConsensusCoreMock) Marshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}

func (cdc *ConsensusCoreMock) MultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}

func (cdc *ConsensusCoreMock) Rounder() consensus.Rounder {
	return cdc.rounder
}

func (cdc *ConsensusCoreMock) ShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}

func (cdc *ConsensusCoreMock) SyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}

func (cdc *ConsensusCoreMock) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}

func (cdc *ConsensusCoreMock) SetBlockchain(blockChain data.ChainHandler) {
	cdc.blockChain = blockChain
}

func (cdc *ConsensusCoreMock) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	cdc.blockProcessor = blockProcessor
}

func (cdc *ConsensusCoreMock) SetBootStrapper(bootstraper process.Bootstrapper) {
	cdc.bootstraper = bootstraper
}

func (cdc *ConsensusCoreMock) SetChronology(chronologyHandler consensus.ChronologyHandler) {
	cdc.chronologyHandler = chronologyHandler
}

func (cdc *ConsensusCoreMock) SetHasher(hasher hashing.Hasher) {
	cdc.hasher = hasher
}

func (cdc *ConsensusCoreMock) SetMarshalizer(marshalizer marshal.Marshalizer) {
	cdc.marshalizer = marshalizer
}

func (cdc *ConsensusCoreMock) SetMultiSigner(multiSigner crypto.MultiSigner) {
	cdc.multiSigner = multiSigner
}

func (cdc *ConsensusCoreMock) SetRounder(rounder consensus.Rounder) {
	cdc.rounder = rounder
}
func (cdc *ConsensusCoreMock) SetShardCoordinator(shardCoordinator sharding.Coordinator) {
	cdc.shardCoordinator = shardCoordinator
}

func (cdc *ConsensusCoreMock) SetSyncTimer(syncTimer ntp.SyncTimer) {
	cdc.syncTimer = syncTimer
}

func (cdc *ConsensusCoreMock) SetValidatorGroupSelector(validatorGroupSelector consensus.ValidatorGroupSelector) {
	cdc.validatorGroupSelector = validatorGroupSelector
}

func (cdc *ConsensusCoreMock) RandomnessPrivateKey() crypto.PrivateKey {
	return cdc.blsPrivateKey
}

// RandomnessSingleSigner returns the bls single signer stored in the ConsensusStore
func (cdc *ConsensusCoreMock) RandomnessSingleSigner() crypto.SingleSigner {
	return cdc.blsSingleSigner
}
