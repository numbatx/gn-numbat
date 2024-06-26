package multisig_test

import (
	"testing"

	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/crypto/mock"
	"github.com/numbatx/gn-numbat/crypto/signing"
	"github.com/numbatx/gn-numbat/crypto/signing/kyber"
	"github.com/numbatx/gn-numbat/crypto/signing/multisig"
	"github.com/numbatx/gn-numbat/hashing"
	"github.com/stretchr/testify/assert"
)

type multiSignerBN interface {
	crypto.MultiSigner
	// CreateCommitment creates a secret commitment and the corresponding public commitment point
	CreateCommitment() (commSecret []byte, commitment []byte)
	// StoreCommitmentHash adds a commitment hash to the list with the specified position
	StoreCommitmentHash(index uint16, commHash []byte) error
	// CommitmentHash returns the commitment hash from the list with the specified position
	CommitmentHash(index uint16) ([]byte, error)
	// StoreCommitment adds a commitment to the list with the specified position
	StoreCommitment(index uint16, value []byte) error
	// Commitment returns the commitment from the list with the specified position
	Commitment(index uint16) ([]byte, error)
	// AggregateCommitments aggregates the list of commitments
	AggregateCommitments(bitmap []byte) error
}

func genMultiSigParams(cnGrSize int, ownIndex uint16) (
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	pubKeys []string,
	kg crypto.KeyGenerator,
) {
	suite := kyber.NewBlakeSHA256Ed25519()
	kg = signing.NewKeyGenerator(suite)

	var pubKeyBytes []byte

	pubKeys = make([]string, 0)
	for i := 0; i < cnGrSize; i++ {
		sk, pk := kg.GeneratePair()
		if uint16(i) == ownIndex {
			privKey = sk
			pubKey = pk
		}

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeys = append(pubKeys, string(pubKeyBytes))
	}

	return privKey, pubKey, pubKeys, kg
}

func setComms(multiSig multiSignerBN, grSize uint16) (bitmap []byte) {
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(0, comm)

	_, comm = multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(grSize-1, comm)

	bitmap = make([]byte, grSize/8+1)
	bitmap[0] = 1
	bitmap[grSize/8] |= 1 << ((grSize - 1) % 8)

	return bitmap
}

func createSignerAndSigShare(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	kg crypto.KeyGenerator,
	grSize uint16,
	ownIndex uint16,
) (sigShare []byte, multiSig multiSignerBN, bitmap []byte) {

	multiSig, _ = multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap = setComms(multiSig, grSize)
	_ = multiSig.AggregateCommitments(bitmap)
	sigShare, _ = multiSig.CreateSignatureShare([]byte("message"), bitmap)

	return sigShare, multiSig, bitmap
}

func createAndSetCommitment(multiSig multiSignerBN, index uint16) (commSecret, comm []byte) {
	commSecret, comm = multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(index, comm)
	return commSecret, comm
}

func createSigShares(
	nbSigs uint16,
	grSize uint16,
	message []byte,
	bitmap []byte,
	ownIndex uint16,
) (sigShares [][]byte, multiSigner multiSignerBN) {

	hasher := &mock.HasherMock{}
	suite := kyber.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)

	var pubKeyBytes []byte

	privKeys := make([]crypto.PrivateKey, grSize)
	pubKeys := make([]crypto.PublicKey, grSize)
	pubKeysStr := make([]string, grSize)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i] = sk
		pubKeys[i] = pk

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	sigShares = make([][]byte, nbSigs)
	multiSigners := make([]multiSignerBN, nbSigs)

	for i := uint16(0); i < nbSigs; i++ {
		multiSigners[i], _ = multisig.NewBelNevMultisig(hasher, pubKeysStr, privKeys[i], kg, i)
	}

	for i := uint16(0); i < nbSigs; i++ {
		_, comm := createAndSetCommitment(multiSigners[i], i)

		// set the <i> commitment for all signers
		for j := uint16(0); j < nbSigs; j++ {
			multiSigners[j].StoreCommitment(i, comm)
		}
	}

	for i := uint16(0); i < nbSigs; i++ {
		_ = multiSigners[i].AggregateCommitments(bitmap)
		sigShares[i], _ = multiSigners[i].CreateSignatureShare(message, bitmap)
	}

	return sigShares, multiSigners[ownIndex]
}

func createAndAddSignatureShares(message []byte) (multiSigner crypto.MultiSigner, bitmap []byte) {
	grSize := uint16(15)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	bitmap = make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	return multiSigner, bitmap
}

func createAggregatedSig(message []byte) (multiSigner crypto.MultiSigner, aggSig []byte, bitmap []byte) {
	multiSigner, bitmap = createAndAddSignatureShares(message)
	aggSig, _ = multiSigner.AggregateSigs(bitmap)

	return multiSigner, aggSig, bitmap
}

func TestNewBelNevMultisig_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(nil, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilHasher, err)
}

func TestNewBelNevMultisig_NilPrivKeyShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	_, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, nil, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestNewBelNevMultisig_NilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, _, kg := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(hasher, nil, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestNewBelNevMultisig_NoPubKeysSetShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, _, kg := genMultiSigParams(4, ownIndex)
	pubKeys := make([]string, 0)

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNoPublicKeySet, err)
}

func TestNewBelNevMultisig_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, _ := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBelNevMultisig_InvalidOwnIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, _ := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBelNevMultisig_OutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, 10)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestNewBelNevMultisig_InvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	pubKeys[1] = "invalid"

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestNewBelNevMultisig_EmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	pubKeys[1] = ""

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestNewBelNevMultisig_OK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, err)
	assert.NotNil(t, multiSig)
}

func TestBelNevSigner_CreateNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	multiSigCreated, err := multiSig.Create(nil, ownIndex)

	assert.Equal(t, crypto.ErrNilPublicKeys, err)
	assert.Nil(t, multiSigCreated)
}

func TestBelNevSigner_CreateInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = "invalid"
	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
	assert.Nil(t, multiSigCreated)
}

func TestBelNevSigner_CreateEmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = ""
	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
	assert.Nil(t, multiSigCreated)
}

func TestBelNevSigner_CreateOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()
	multiSig.StoreCommitment(0, comm)

	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)
	mSig, _ := multiSigCreated.(multiSignerBN)

	assert.Nil(t, err)
	assert.NotNil(t, mSig)

	_, err = mSig.Commitment(ownIndex)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_ResetOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.Reset(pubKeys, 10)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_ResetNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.Reset(nil, ownIndex)

	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestBelNevSigner_ResetInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = "invalid"
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestBelNevSigner_ResetEmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = ""
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestBelNevSigner_ResetOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.Reset(pubKeys, ownIndex)
	assert.Nil(t, err)
}

func TestBelNevSigner_AddCommitmentHashInvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	// invalid index > len(pubKeys)
	err := multiSig.StoreCommitmentHash(100, commHash)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_AddCommitmentHashNilCommitmentHashShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.StoreCommitmentHash(0, nil)

	assert.Equal(t, crypto.ErrNilCommitmentHash, err)
}

func TestBelNevSigner_AddCommitmentHashOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	err := multiSig.StoreCommitmentHash(0, commHash)
	assert.Nil(t, err)

	commHashRead, _ := multiSig.CommitmentHash(0)
	assert.Equal(t, commHash, commHashRead)
}

func TestBelNevSigner_CommitmentHashIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	_ = multiSig.StoreCommitmentHash(0, commHash)
	// index > len(pubKeys)
	commHashRead, err := multiSig.CommitmentHash(100)

	assert.Nil(t, commHashRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_CommitmentHashNotSetIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHashRead, err := multiSig.CommitmentHash(0)

	assert.Nil(t, commHashRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_CommitmentHashOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	_ = multiSig.StoreCommitmentHash(0, commHash)
	commHashRead, err := multiSig.CommitmentHash(0)

	assert.Nil(t, err)
	assert.Equal(t, commHash, commHashRead)
}

func TestBelNevSigner_CreateCommitmentOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	secret, comm := multiSig.CreateCommitment()

	sk, _ := kg.PrivateKeyFromByteArray(secret)
	pk := sk.GeneratePublic()
	pkBytes, err := pk.Point().MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.NotNil(t, comm)
	assert.Equal(t, comm, pkBytes)
}

func TestBelNevSigner_AddCommitmentNilCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.StoreCommitment(0, nil)

	assert.Equal(t, crypto.ErrNilCommitment, err)
}

func TestBelNevSigner_AddCommitmentIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	err := multiSig.StoreCommitment(100, commitment)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_AddCommitmentOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	err := multiSig.StoreCommitment(0, commitment)

	assert.Nil(t, err)
}

func TestBelNevSigner_CommitmentIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(0, commitment)

	commRead, err := multiSig.Commitment(100)

	assert.Nil(t, commRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_CommitmentNilCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commRead, err := multiSig.Commitment(0)

	assert.Nil(t, commRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_CommitmentOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(0, commitment)
	commRead, err := multiSig.Commitment(0)

	assert.Nil(t, err)
	assert.Equal(t, commRead, commitment)
}

func TestBelNevSigner_CommitmentSecretNotSetShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commSecretRead, err := multiSig.CommitmentSecret()

	assert.Nil(t, commSecretRead)
	assert.Equal(t, crypto.ErrNilCommitmentSecret, err)
}

func TestBelNevSigner_CommitmentSecretOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commSecret, _ := multiSig.CreateCommitment()
	commSecretRead, err := multiSig.CommitmentSecret()

	assert.Nil(t, err)
	assert.Equal(t, commSecret, commSecretRead)
}

func TestBelNevSigner_AggregateCommitmentsNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_ = setComms(multiSig, 4)
	err := multiSig.AggregateCommitments(nil)

	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_AggregateCommitmentsWrongBitmapShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(20)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(21, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 21)

	err := multiSig.AggregateCommitments(bitmap[:1])

	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_AggregateCommitmentsNotSetCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(3, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(0, commitment)
	_, commitment2 := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(2, commitment2)

	bitmap := make([]byte, 1)
	bitmap[0] |= 7 // 0b00000111

	err := multiSig.AggregateCommitments(bitmap)

	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestBelNevSigner_AggregateCommitmentsOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(3, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	err := multiSig.AggregateCommitments(bitmap)

	assert.Nil(t, err)
}

func TestBelNevSigner_CreateSignatureShareNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(2)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(3, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 3)
	_ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare([]byte("message"), nil)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_CreateSignatureShareInvalidBitmapShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(20)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(21, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 21)
	_ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare([]byte("message"), bitmap[:1])

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_CreateSignatureShareNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	_ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare(nil, bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetAggCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	sigShare, err := multiSig.CreateSignatureShare([]byte("message"), bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilAggregatedCommitment, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetCommSecretShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.StoreCommitment(ownIndex, comm)

	bitmap := make([]byte, 1)
	bitmap[0] = 1 << 3

	_ = multiSig.AggregateCommitments(bitmap)

	multiSigCreated, _ := multiSig.Create(pubKeys, ownIndex)
	ms, _ := multiSigCreated.(multiSignerBN)

	_ = ms.StoreCommitment(ownIndex, comm)
	_ = ms.AggregateCommitments(bitmap)

	sigShare, err := ms.CreateSignatureShare([]byte("message"), bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilCommitmentSecret, err)
}

func TestBelNevSigner_CreateSignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	_ = multiSig.AggregateCommitments(bitmap)
	msg := []byte("message")
	sigShare, err := multiSig.CreateSignatureShare(msg, bitmap)

	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare, msg, bitmap)

	assert.Nil(t, err)
	assert.NotNil(t, sigShare)
	assert.Nil(t, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	_, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	verifErr := multiSig.VerifySignatureShare(ownIndex, nil, msg, bitmap)

	assert.Equal(t, crypto.ErrNilSignature, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare, msg, nil)

	assert.Equal(t, crypto.ErrNilBitmap, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNotSelectedIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	bitmap[0] = 0
	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare, msg, bitmap)

	assert.Equal(t, crypto.ErrIndexNotSelected, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNotSetCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	bitmap[0] |= 1 << 2
	verifErr := multiSig.VerifySignatureShare(2, sigShare, msg, bitmap)

	assert.Equal(t, crypto.ErrNilParam, verifErr)
}

func TestBelNevSigner_VerifySignatureShareInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	verifErr := multiSig.VerifySignatureShare(0, sigShare, msg, bitmap)

	assert.Equal(t, crypto.ErrSigNotValid, verifErr)
}

func TestBelNevSigner_VerifySignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	msg := []byte("message")

	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare, msg, bitmap)

	assert.Nil(t, verifErr)
}

func TestBelNevSigner_AddSignatureShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.StoreSignatureShare(ownIndex, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBelNevSigner_AddSignatureShareInvalidSigScalarShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	sig := []byte("invalid signature")

	err := multiSig.StoreSignatureShare(ownIndex, sig)

	assert.NotNil(t, err)
}

func TestBelNevSigner_AddSignatureShareIndexOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	bitmap := setComms(multiSig, 4)
	_ = multiSig.AggregateCommitments(bitmap)
	msg := []byte("message")
	sigShare, _ := multiSig.CreateSignatureShare(msg, bitmap)

	err := multiSig.StoreSignatureShare(15, sigShare)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_AddSignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	err := multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, _ := multiSig.SignatureShare(ownIndex)

	assert.Nil(t, err)
	assert.Equal(t, sigShare, sigShareRead)
}

func TestBelNevSigner_SignatureShareOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(15)

	assert.Nil(t, sigShareRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_SignatureShareNotSetIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(2)

	assert.Nil(t, sigShareRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_SignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg := genMultiSigParams(4, ownIndex)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(ownIndex)

	assert.Nil(t, err)
	assert.Equal(t, sigShare, sigShareRead)
}

func TestBelNevSigner_AggregateSigsNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(nil)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_AggregateSigsInvalidBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(21)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 3)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	bitmap = make([]byte, 1)
	bitmap[0] = 0x07

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_AggregateSigsMissingSigShareShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs)-1; i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestBelNevSigner_AggregateSigsZeroSelectionBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs)-1; i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}
	bitmap[0] = 0
	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrBitmapNotSet, err)
}

func TestBelNevSigner_AggregateSigsOK(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigShares(nbSigners, grSize, message, bitmap, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, err)
	assert.NotNil(t, aggSig)
}

func TestBelNevSigner_SetAggregatedSigNilSigShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, _, _ := createAggregatedSig(msg)
	err := multiSigner.SetAggregatedSig(nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBelNevSigner_SetAggregatedSigInvalidScalarShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, _, _ := createAggregatedSig(msg)
	aggSig := []byte("invalid agg signature xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	err := multiSigner.SetAggregatedSig(aggSig)

	assert.Equal(t, err, crypto.ErrAggSigNotValid)
}

func TestBelNevSigner_SetAggregatedSigOK(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, aggSig, _ := createAggregatedSig(msg)
	err := multiSigner.SetAggregatedSig(aggSig)

	assert.Nil(t, err)
}

func TestBelNevSigner_VerifyNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, aggSig, _ := createAggregatedSig(msg)
	_ = multiSigner.SetAggregatedSig(aggSig)
	err := multiSigner.Verify(msg, nil)

	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_VerifyBitmapMismatchShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, aggSig, bitmap := createAggregatedSig(msg)
	_ = multiSigner.SetAggregatedSig(aggSig)
	// set a smaller bitmap
	bitmap = make([]byte, 1)

	err := multiSigner.Verify(msg, bitmap)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_VerifyAggSigNotSetShouldErr(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, bitmap := createAndAddSignatureShares(msg)
	err := multiSigner.Verify(msg, bitmap)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBelNevSigner_VerifySigValid(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, aggSig, bitmap := createAggregatedSig(msg)
	_ = multiSigner.SetAggregatedSig(aggSig)

	err := multiSigner.Verify(msg, bitmap)
	assert.Nil(t, err)
}

func TestBelNevSigner_VerifySigInvalid(t *testing.T) {
	t.Parallel()

	msg := []byte("message")
	multiSigner, aggSig, bitmap := createAggregatedSig(msg)
	// make sig invalid
	aggSig[len(aggSig)-1] = aggSig[len(aggSig)-1] ^ 255
	_ = multiSigner.SetAggregatedSig(aggSig)

	err := multiSigner.Verify(msg, bitmap)
	assert.Equal(t, crypto.ErrSigNotValid, err)
}
