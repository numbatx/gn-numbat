package transaction

import (
	"io"
	"math/big"

	capn "github.com/glycerine/go-capnproto"
	"github.com/numbatx/gn-numbat/data/transaction/capnp"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     uint64   `capid:"0"`
	Value     *big.Int `capid:"1"`
	RcvAddr   []byte   `capid:"2"`
	SndAddr   []byte   `capid:"3"`
	GasPrice  uint64   `capid:"4"`
	GasLimit  uint64   `capid:"5"`
	Data      []byte   `capid:"6"`
	Signature []byte   `capid:"7"`
	Challenge []byte   `capid:"8"`
}

// Save saves the serialized data of a Transaction into a stream through Capnp protocol
func (tx *Transaction) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	TransactionGoToCapn(seg, tx)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Transaction object through Capnp protocol
func (tx *Transaction) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootTransactionCapn(capMsg)
	TransactionCapnToGo(z, tx)
	return nil
}

// TransactionCapnToGo is a helper function to copy fields from a TransactionCapn object to a Transaction object
func TransactionCapnToGo(src capnp.TransactionCapn, dest *Transaction) *Transaction {
	if dest == nil {
		dest = &Transaction{}
	}

	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}

	// Nonce
	dest.Nonce = src.Nonce()
	// Value
	err := dest.Value.GobDecode(src.Value())

	if err != nil {
		return nil
	}

	// RcvAddr
	dest.RcvAddr = src.RcvAddr()
	// SndAddr
	dest.SndAddr = src.SndAddr()
	// GasPrice
	dest.GasPrice = src.GasPrice()
	// GasLimit
	dest.GasLimit = src.GasLimit()
	// Data
	dest.Data = src.Data()
	// Signature
	dest.Signature = src.Signature()
	// Challenge
	dest.Challenge = src.Challenge()

	return dest
}

// TransactionGoToCapn is a helper function to copy fields from a Transaction object to a TransactionCapn object
func TransactionGoToCapn(seg *capn.Segment, src *Transaction) capnp.TransactionCapn {
	dest := capnp.AutoNewTransactionCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetNonce(src.Nonce)
	dest.SetValue(value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetSndAddr(src.SndAddr)
	dest.SetGasPrice(src.GasPrice)
	dest.SetGasLimit(src.GasLimit)
	dest.SetData(src.Data)
	dest.SetSignature(src.Signature)
	dest.SetChallenge(src.Challenge)

	return dest
}
