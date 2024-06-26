package block

import (
	"bytes"
	"fmt"

	"github.com/numbatx/gn-numbat/core"
	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/data"
	"github.com/numbatx/gn-numbat/data/state"
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/display"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/numbatx/gn-numbat/marshal"
	"github.com/numbatx/gn-numbat/process"
	"github.com/numbatx/gn-numbat/sharding"
)

var log = logger.DefaultLogger()

type baseProcessor struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	forkDetector     process.ForkDetector
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService
}

func checkForNils(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	if chainHandler == nil {
		return process.ErrNilBlockChain
	}
	if headerHandler == nil {
		return process.ErrNilBlockHeader
	}
	if bodyHandler == nil {
		return process.ErrNilBlockBody
	}
	return nil
}

// RevertAccountState reverts the account state for cleanup failed process
func (bp *baseProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
}

// checkBlockValidity method checks if the given block is valid
func (bp *baseProcessor) checkBlockValidity(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	err := checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	if chainHandler.GetCurrentBlockHeader() == nil {
		if headerHandler.GetNonce() == 1 { // first block after genesis
			if bytes.Equal(headerHandler.GetPrevHash(), chainHandler.GetGenesisHeaderHash()) {
				// TODO: add genesis block verification
				return nil
			}

			log.Info(fmt.Sprintf("hash not match: local block hash is empty and node received block with previous hash %s\n",
				core.ToB64(headerHandler.GetPrevHash())))

			return process.ErrInvalidBlockHash
		}

		log.Info(fmt.Sprintf("nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetNonce() != chainHandler.GetCurrentBlockHeader().GetNonce()+1 {
		log.Info(fmt.Sprintf("nonce not match: local block nonce is %d and node received block with nonce %d\n",
			chainHandler.GetCurrentBlockHeader().GetNonce(), headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := bp.computeHeaderHash(chainHandler.GetCurrentBlockHeader())
	if err != nil {
		return err
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), prevHeaderHash) {
		log.Info(fmt.Sprintf("hash not match: local block hash is %s and node received block with previous hash %s\n",
			core.ToB64(prevHeaderHash), core.ToB64(headerHandler.GetPrevHash())))

		return process.ErrInvalidBlockHash
	}

	if bodyHandler != nil {
		// TODO: add bodyHandler verification here
	}

	// TODO: add signature validation as well, with randomness source and all
	return nil
}

// verifyStateRoot verifies the state root hash given as parameter against the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *baseProcessor) verifyStateRoot(rootHash []byte) bool {
	return bytes.Equal(bp.accounts.RootHash(), rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	return bp.accounts.RootHash()
}

func (bp *baseProcessor) computeHeaderHash(headerHandler data.HeaderHandler) ([]byte, error) {
	headerMarsh, err := bp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return nil, err
	}

	headerHash := bp.hasher.Compute(string(headerMarsh))

	return headerHash, nil
}

func displayHeader(headerHandler data.HeaderHandler) []*display.LineData {
	lines := make([]*display.LineData, 0)

	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Epoch",
		fmt.Sprintf("%d", headerHandler.GetEpoch())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Round",
		fmt.Sprintf("%d", headerHandler.GetRound())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"TimeStamp",
		fmt.Sprintf("%d", headerHandler.GetTimeStamp())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Nonce",
		fmt.Sprintf("%d", headerHandler.GetNonce())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev hash",
		core.ToB64(headerHandler.GetPrevHash())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev rand seed",
		core.ToB64(headerHandler.GetPrevRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Rand seed",
		core.ToB64(headerHandler.GetRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		core.ToHex(headerHandler.GetPubKeysBitmap())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Signature",
		core.ToB64(headerHandler.GetSignature())}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Root hash",
		core.ToB64(headerHandler.GetRootHash())}))
	return lines
}

// checkProcessorNilParameters will check the imput parameters for nil values
func checkProcessorNilParameters(
	accounts state.AccountsAdapter,
	forkDetector process.ForkDetector,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
	shardCoordinator sharding.Coordinator,
) error {

	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}
	if forkDetector == nil {
		return process.ErrNilForkDetector
	}
	if hasher == nil {
		return process.ErrNilHasher
	}
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if store == nil {
		return process.ErrNilStorage
	}
	if shardCoordinator == nil {
		return process.ErrNilShardCoordinator
	}

	return nil
}
