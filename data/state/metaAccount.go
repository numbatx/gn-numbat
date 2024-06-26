package state

import (
	"math/big"

	"github.com/numbatx/gn-numbat/data/trie"
)

// MiniBlockData is the data to be saved in shard account for any shard
type MiniBlockData struct {
	Hash            []byte
	ReceiverShardId uint32
	SenderShardId   uint32
	TxCount         uint32
}

// MetaAccount is the struct used in serialization/deserialization
type MetaAccount struct {
	Round         uint64
	TxCount       *big.Int
	CodeHash      []byte
	RootHash      []byte
	MiniBlocks    []*MiniBlockData
	PubKeyLeader  []byte
	ShardRootHash []byte

	addressContainer AddressContainer
	code             []byte
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewMetaAccount creates new simple meta account for an AccountContainer (that has just been initialized)
func NewMetaAccount(addressContainer AddressContainer, tracker AccountTracker) (*MetaAccount, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}
	if tracker == nil {
		return nil, ErrNilAccountTracker
	}

	return &MetaAccount{
		TxCount:          big.NewInt(0),
		addressContainer: addressContainer,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(nil),
	}, nil
}

// AddressContainer returns the address associated with the account
func (a *MetaAccount) AddressContainer() AddressContainer {
	return a.addressContainer
}

// SetRoundWithJournal sets the account's round, saving the old round before changing
func (a *MetaAccount) SetRoundWithJournal(round uint64) error {
	entry, err := NewMetaJournalEntryRound(a, a.Round)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Round = round

	return a.accountTracker.SaveAccount(a)
}

// SetTxCountWithJournal sets the total tx count for this shard, saving the old txCount before changing
func (a *MetaAccount) SetTxCountWithJournal(txCount *big.Int) error {
	entry, err := NewMetaJournalEntryTxCount(a, a.TxCount)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.TxCount = txCount

	return a.accountTracker.SaveAccount(a)
}

// SetMiniBlocksDataWithJournal sets the current final mini blocks header data,
// saving the old mini blocks header data before changing
func (a *MetaAccount) SetMiniBlocksDataWithJournal(miniBlocksData []*MiniBlockData) error {
	entry, err := NewMetaJournalEntryMiniBlocksData(a, a.MiniBlocks)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.MiniBlocks = miniBlocksData

	return a.accountTracker.SaveAccount(a)
}

// SetShardRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (a *MetaAccount) SetShardRootHashWithJournal(shardRootHash []byte) error {
	entry, err := NewMetaJournalEntryShardRootHash(a, a.ShardRootHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.ShardRootHash = shardRootHash

	return a.accountTracker.SaveAccount(a)
}

//------- code / code hash

// GetCodeHash returns the code hash associated with this account
func (a *MetaAccount) GetCodeHash() []byte {
	return a.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (a *MetaAccount) SetCodeHash(roothash []byte) {
	a.CodeHash = roothash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (a *MetaAccount) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(a, a.CodeHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.CodeHash = codeHash

	return a.accountTracker.SaveAccount(a)
}

// GetCode gets the actual code that needs to be run in the VM
func (a *MetaAccount) GetCode() []byte {
	return a.code
}

// SetCode sets the actual code that needs to be run in the VM
func (a *MetaAccount) SetCode(code []byte) {
	a.code = code
}

//------- data trie / root hash

// GetRootHash returns the root hash associated with this account
func (a *MetaAccount) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *MetaAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// SetRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (a *MetaAccount) SetRootHashWithJournal(rootHash []byte) error {
	entry, err := NewBaseJournalEntryRootHash(a, a.RootHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.RootHash = rootHash

	return a.accountTracker.SaveAccount(a)
}

// DataTrie returns the trie that holds the current account's data
func (a *MetaAccount) DataTrie() trie.PatriciaMerkelTree {
	return a.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (a *MetaAccount) SetDataTrie(trie trie.PatriciaMerkelTree) {
	a.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (a *MetaAccount) DataTrieTracker() DataTrieTracker {
	return a.dataTrieTracker
}

//TODO add Cap'N'Proto converter funcs
