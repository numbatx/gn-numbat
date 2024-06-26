package spos

import (
	"github.com/numbatx/gn-numbat/consensus"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process"
)

type RoundConsensus *roundConsensus

// worker

func (wrk *Worker) BlockProcessor() process.BlockProcessor {
	return wrk.blockProcessor
}

func (wrk *Worker) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	wrk.blockProcessor = blockProcessor
}

func (wrk *Worker) BlockTracker() process.BlocksTracker {
	return wrk.blockTracker
}

func (wrk *Worker) SetBlockTracker(blockTracker process.BlocksTracker) {
	wrk.blockTracker = blockTracker
}

func (wrk *Worker) Bootstraper() process.Bootstrapper {
	return wrk.bootstraper
}

func (wrk *Worker) SetBootstraper(bootstraper process.Bootstrapper) {
	wrk.bootstraper = bootstraper
}

func (wrk *Worker) ConsensusState() *ConsensusState {
	return wrk.consensusState
}

func (wrk *Worker) SetConsensusState(consensusState *ConsensusState) {
	wrk.consensusState = consensusState
}

func (wrk *Worker) ForkDetector() process.ForkDetector {
	return wrk.forkDetector
}

func (wrk *Worker) SetForkDetector(forkDetector process.ForkDetector) {
	wrk.forkDetector = forkDetector
}

func (wrk *Worker) KeyGenerator() crypto.KeyGenerator {
	return wrk.keyGenerator
}

func (wrk *Worker) SetKeyGenerator(keyGenerator crypto.KeyGenerator) {
	wrk.keyGenerator = keyGenerator
}

func (wrk *Worker) Marshalizer() marshal.Marshalizer {
	return wrk.marshalizer
}

func (wrk *Worker) SetMarshalizer(marshalizer marshal.Marshalizer) {
	wrk.marshalizer = marshalizer
}

func (wrk *Worker) Rounder() consensus.Rounder {
	return wrk.rounder
}

func (wrk *Worker) SetRounder(rounder consensus.Rounder) {
	wrk.rounder = rounder
}

func (wrk *Worker) CheckSignature(cnsData *consensus.Message) error {
	return wrk.checkSignature(cnsData)
}

func (wrk *Worker) ExecuteMessage(cnsDtaList []*consensus.Message) {
	wrk.executeMessage(cnsDtaList)
}

func (wrk *Worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *Worker) ReceivedSyncState(isNodeSynchronized bool) {
	wrk.receivedSyncState(isNodeSynchronized)
}

func (wrk *Worker) ReceivedMessages() map[consensus.MessageType][]*consensus.Message {
	wrk.mutReceivedMessages.RLock()
	defer wrk.mutReceivedMessages.RUnlock()

	return wrk.receivedMessages
}

func (wrk *Worker) SetReceivedMessages(messageType consensus.MessageType, cnsDta []*consensus.Message) {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages[messageType] = cnsDta
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) NilReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = nil
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) ReceivedMessagesCalls() map[consensus.MessageType]func(*consensus.Message) bool {
	wrk.mutReceivedMessagesCalls.RLock()
	defer wrk.mutReceivedMessagesCalls.RUnlock()

	return wrk.receivedMessagesCalls
}

func (wrk *Worker) SetReceivedMessagesCalls(messageType consensus.MessageType, f func(*consensus.Message) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = f
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *Worker) ExecuteMessageChannel() chan *consensus.Message {
	return wrk.executeMessageChannel
}

func (wrk *Worker) ConsensusStateChangedChannel() chan bool {
	return wrk.consensusStateChangedChannel
}

func (wrk *Worker) SetConsensusStateChangedChannel(consensusStateChangedChannel chan bool) {
	wrk.consensusStateChangedChannel = consensusStateChangedChannel
}

func (wrk *Worker) CheckSelfState(cnsDta *consensus.Message) error {
	return wrk.checkSelfState(cnsDta)
}

func (wrk *Worker) SetSendMessage(sendMessage func(consensus *consensus.Message)) {
	wrk.sendMessage = sendMessage
}

func (wrk *Worker) SetBroadcastBlock(broadcastBlock func(data.BodyHandler, data.HeaderHandler) error) {
	wrk.broadcastBlock = broadcastBlock
}

func (wrk *Worker) SetBroadcastHeader(broadcastHeader func(data.HeaderHandler) error) {
	wrk.broadcastHeader = broadcastHeader
}
