package spos

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/numbatx/gn-numbat/consensus"
	"github.com/numbatx/gn-numbat/core"
	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/ntp"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/sharding"
)

// Worker defines the data needed by spos to communicate between nodes which are in the validators group
type Worker struct {
	consensusService ConsensusService
	blockProcessor   process.BlockProcessor
	blockTracker     process.BlocksTracker
	bootstraper      process.Bootstrapper
	consensusState   *ConsensusState
	forkDetector     process.ForkDetector
	keyGenerator     crypto.KeyGenerator
	marshalizer      marshal.Marshalizer
	privateKey       crypto.PrivateKey
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator
	singleSigner     crypto.SingleSigner
	syncTimer        ntp.SyncTimer

	receivedMessages      map[consensus.MessageType][]*consensus.Message
	receivedMessagesCalls map[consensus.MessageType]func(*consensus.Message) bool

	executeMessageChannel        chan *consensus.Message
	consensusStateChangedChannel chan bool

	broadcastBlock  func(data.BodyHandler, data.HeaderHandler) error
	broadcastHeader func(data.HeaderHandler) error
	sendMessage     func(consensus *consensus.Message)

	mutReceivedMessages      sync.RWMutex
	mutReceivedMessagesCalls sync.RWMutex
}

// NewWorker creates a new Worker object
func NewWorker(
	consensusService ConsensusService,
	blockProcessor process.BlockProcessor,
	blockTracker process.BlocksTracker,
	bootstraper process.Bootstrapper,
	consensusState *ConsensusState,
	forkDetector process.ForkDetector,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
	syncTimer ntp.SyncTimer,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
	broadcastHeader func(data.HeaderHandler) error,
	sendMessage func(consensus *consensus.Message),
) (*Worker, error) {
	err := checkNewWorkerParams(
		consensusService,
		blockProcessor,
		blockTracker,
		bootstraper,
		consensusState,
		forkDetector,
		keyGenerator,
		marshalizer,
		privateKey,
		rounder,
		shardCoordinator,
		singleSigner,
		syncTimer,
		broadcastBlock,
		broadcastHeader,
		sendMessage,
	)
	if err != nil {
		return nil, err
	}

	wrk := Worker{
		consensusService: consensusService,
		blockProcessor:   blockProcessor,
		blockTracker:     blockTracker,
		bootstraper:      bootstraper,
		consensusState:   consensusState,
		forkDetector:     forkDetector,
		keyGenerator:     keyGenerator,
		marshalizer:      marshalizer,
		privateKey:       privateKey,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
		syncTimer:        syncTimer,
		broadcastBlock:   broadcastBlock,
		broadcastHeader:  broadcastHeader,
		sendMessage:      sendMessage,
	}

	wrk.executeMessageChannel = make(chan *consensus.Message)
	wrk.receivedMessagesCalls = make(map[consensus.MessageType]func(*consensus.Message) bool)
	wrk.consensusStateChangedChannel = make(chan bool, 1)
	wrk.bootstraper.AddSyncStateListener(wrk.receivedSyncState)
	wrk.initReceivedMessages()

	go wrk.checkChannels()

	return &wrk, nil
}

func checkNewWorkerParams(
	consensusService ConsensusService,
	blockProcessor process.BlockProcessor,
	blockTracker process.BlocksTracker,
	bootstraper process.Bootstrapper,
	consensusState *ConsensusState,
	forkDetector process.ForkDetector,
	keyGenerator crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	privateKey crypto.PrivateKey,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
	syncTimer ntp.SyncTimer,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
	broadcastHeader func(data.HeaderHandler) error,
	sendMessage func(consensus *consensus.Message),
) error {
	if consensusService == nil {
		return ErrNilConsensusService
	}
	if blockProcessor == nil {
		return ErrNilBlockProcessor
	}
	if blockTracker == nil {
		return ErrNilBlockTracker
	}
	if bootstraper == nil {
		return ErrNilBlootstraper
	}
	if consensusState == nil {
		return ErrNilConsensusState
	}
	if forkDetector == nil {
		return ErrNilForkDetector
	}
	if keyGenerator == nil {
		return ErrNilKeyGenerator
	}
	if marshalizer == nil {
		return ErrNilMarshalizer
	}
	if privateKey == nil {
		return ErrNilPrivateKey
	}
	if rounder == nil {
		return ErrNilRounder
	}
	if shardCoordinator == nil {
		return ErrNilShardCoordinator
	}
	if singleSigner == nil {
		return ErrNilSingleSigner
	}
	if syncTimer == nil {
		return ErrNilSyncTimer
	}
	if broadcastBlock == nil {
		return ErrNilBroadcastBlock
	}
	if broadcastHeader == nil {
		return ErrNilBroadcastHeader
	}
	if sendMessage == nil {
		return ErrNilSendMessage
	}

	return nil
}

func (wrk *Worker) receivedSyncState(isNodeSynchronized bool) {
	if isNodeSynchronized {
		if len(wrk.consensusStateChangedChannel) == 0 {
			wrk.consensusStateChangedChannel <- true
		}
	}
}

func (wrk *Worker) initReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = wrk.consensusService.InitReceivedMessages()
	wrk.mutReceivedMessages.Unlock()
}

// AddReceivedMessageCall adds a new handler function for a received messege type
func (wrk *Worker) AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = receivedMessageCall
	wrk.mutReceivedMessagesCalls.Unlock()
}

// RemoveAllReceivedMessagesCalls removes all the functions handlers
func (wrk *Worker) RemoveAllReceivedMessagesCalls() {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls = make(map[consensus.MessageType]func(*consensus.Message) bool)
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *Worker) getCleanedList(cnsDataList []*consensus.Message) []*consensus.Message {
	cleanedCnsDataList := make([]*consensus.Message, 0)

	for i := 0; i < len(cnsDataList); i++ {
		if cnsDataList[i] == nil {
			continue
		}

		if wrk.rounder.Index() > cnsDataList[i].RoundIndex {
			continue
		}

		cleanedCnsDataList = append(cleanedCnsDataList, cnsDataList[i])
	}

	return cleanedCnsDataList
}

// ProcessReceivedMessage method redirects the received message to the channel which should handle it
func (wrk *Worker) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
		return ErrNilMessage
	}

	if message.Data() == nil {
		return ErrNilDataToProcess
	}

	cnsDta := &consensus.Message{}
	err := wrk.marshalizer.Unmarshal(cnsDta, message.Data())
	if err != nil {
		return err
	}

	msgType := consensus.MessageType(cnsDta.MsgType)

	log.Debug(fmt.Sprintf("received %s from %s for consensus message with with header hash %s and round %d\n",
		wrk.consensusService.GetStringValue(msgType),
		core.GetTrimmedPk(hex.EncodeToString(cnsDta.PubKey)),
		base64.StdEncoding.EncodeToString(cnsDta.BlockHeaderHash),
		cnsDta.RoundIndex,
	))

	senderOK := wrk.consensusState.IsNodeInEligibleList(string(cnsDta.PubKey))
	if !senderOK {
		return ErrSenderNotOk
	}

	if wrk.consensusState.RoundIndex > cnsDta.RoundIndex {
		return ErrMessageForPastRound
	}

	sigVerifErr := wrk.checkSignature(cnsDta)
	if sigVerifErr != nil {
		return ErrInvalidSignature
	}

	if wrk.consensusService.IsMessageWithBlockHeader(msgType) {
		headerHash := cnsDta.BlockHeaderHash
		header := wrk.blockProcessor.DecodeBlockHeader(cnsDta.SubRoundData)
		errNotCritical := wrk.forkDetector.AddHeader(header, headerHash, process.BHProposed)
		if errNotCritical != nil {
			log.Debug(errNotCritical.Error())
		}
	}

	errNotCritical := wrk.checkSelfState(cnsDta)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
		//in this case should return nil but do not process the message
		//nil error will mean that the interceptor will validate this message and broadcast it to the connected peers
		return nil
	}

	go wrk.executeReceivedMessages(cnsDta)

	return nil
}

func (wrk *Worker) checkSelfState(cnsDta *consensus.Message) error {
	if wrk.consensusState.SelfPubKey() == string(cnsDta.PubKey) {
		return ErrMessageFromItself
	}

	if wrk.consensusState.RoundCanceled && wrk.consensusState.RoundIndex == cnsDta.RoundIndex {
		return ErrRoundCanceled
	}

	return nil
}

func (wrk *Worker) checkSignature(cnsDta *consensus.Message) error {
	if cnsDta == nil {
		return ErrNilConsensusData
	}
	if cnsDta.PubKey == nil {
		return ErrNilPublicKey
	}
	if cnsDta.Signature == nil {
		return ErrNilSignature
	}

	pubKey, err := wrk.keyGenerator.PublicKeyFromByteArray(cnsDta.PubKey)
	if err != nil {
		return err
	}

	dataNoSig := *cnsDta
	signature := cnsDta.Signature
	dataNoSig.Signature = nil
	dataNoSigString, err := wrk.marshalizer.Marshal(dataNoSig)
	if err != nil {
		return err
	}

	err = wrk.singleSigner.Verify(pubKey, dataNoSigString, signature)
	return err
}

func (wrk *Worker) executeReceivedMessages(cnsDta *consensus.Message) {
	wrk.mutReceivedMessages.Lock()

	msgType := consensus.MessageType(cnsDta.MsgType)
	cnsDataList := wrk.receivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.receivedMessages[msgType] = cnsDataList
	wrk.executeStoredMessages()

	wrk.mutReceivedMessages.Unlock()
}

func (wrk *Worker) executeStoredMessages() {
	for _, i := range wrk.consensusService.GetMessageRange() {
		cnsDataList := wrk.receivedMessages[i]
		if len(cnsDataList) == 0 {
			continue
		}
		wrk.executeMessage(cnsDataList)
		cleanedCnsDtaList := wrk.getCleanedList(cnsDataList)
		wrk.receivedMessages[i] = cleanedCnsDtaList
	}
}

func (wrk *Worker) executeMessage(cnsDtaList []*consensus.Message) {
	for i, cnsDta := range cnsDtaList {
		if cnsDta == nil {
			continue
		}
		if wrk.consensusState.RoundIndex != cnsDta.RoundIndex {
			continue
		}

		msgType := consensus.MessageType(cnsDta.MsgType)
		if !wrk.consensusService.CanProceed(wrk.consensusState, msgType) {
			continue
		}

		cnsDtaList[i] = nil
		wrk.executeMessageChannel <- cnsDta
	}
}

// checkChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (wrk *Worker) checkChannels() {
	for {
		select {
		case rcvDta := <-wrk.executeMessageChannel:
			msgType := consensus.MessageType(rcvDta.MsgType)
			if callReceivedMessage, exist := wrk.receivedMessagesCalls[msgType]; exist {
				if callReceivedMessage(rcvDta) {
					if len(wrk.consensusStateChangedChannel) == 0 {
						wrk.consensusStateChangedChannel <- true
					}
				}
			}
		}
	}
}

// SendConsensusMessage sends the consensus message
func (wrk *Worker) SendConsensusMessage(cnsDta *consensus.Message) bool {
	signature, err := wrk.genConsensusDataSignature(cnsDta)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	signedCnsData := *cnsDta
	signedCnsData.Signature = signature

	if wrk.sendMessage == nil {
		log.Error("sendMessage call back function is not set\n")
		return false
	}

	go wrk.sendMessage(&signedCnsData)

	return true
}

func (wrk *Worker) genConsensusDataSignature(cnsDta *consensus.Message) ([]byte, error) {
	cnsDtaStr, err := wrk.marshalizer.Marshal(cnsDta)
	if err != nil {
		return nil, err
	}

	signature, err := wrk.singleSigner.Sign(wrk.privateKey, cnsDtaStr)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Extend does an extension for the subround with subroundId
func (wrk *Worker) Extend(subroundId int) {
	log.Info(fmt.Sprintf("extend function is called from subround: %s\n",
		wrk.consensusService.GetSubroundName(subroundId)))

	if wrk.bootstraper.ShouldSync() {
		return
	}

	for wrk.consensusState.ProcessingBlock() {
		time.Sleep(time.Millisecond)
	}

	wrk.blockProcessor.RevertAccountState()
}

// GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
func (wrk *Worker) GetConsensusStateChangedChannel() chan bool {
	return wrk.consensusStateChangedChannel
}

// BroadcastBlock does a broadcast of the blockBody and blockHeader
func (wrk *Worker) BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error {
	return wrk.broadcastBlock(body, header)
}

// BroadcastUnnotarisedBlocks broadcasts all blocks which are not notarised yet
func (wrk *Worker) BroadcastUnnotarisedBlocks() {
	headers := wrk.blockTracker.UnnotarisedBlocks()
	for _, header := range headers {
		if header.GetNonce() > wrk.forkDetector.GetHighestFinalBlockNonce() {
			continue
		}

		brodcastRound := wrk.blockTracker.BlockBroadcastRound(header.GetNonce())
		if brodcastRound >= wrk.consensusState.RoundIndex-MaxRoundsGap {
			continue
		}

		err := wrk.broadcastHeader(header)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		wrk.blockTracker.SetBlockBroadcastRound(header.GetNonce(), wrk.consensusState.RoundIndex)

		log.Info(fmt.Sprintf("%sStep 0: Unnotarised header with nonce %d has been broadcast to metachain\n",
			wrk.syncTimer.FormattedCurrentTime(),
			header.GetNonce()))
	}
}

// ExecuteStoredMessages tries to execute all the messages received which are valid for execution
func (wrk *Worker) ExecuteStoredMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.executeStoredMessages()
	wrk.mutReceivedMessages.Unlock()
}
