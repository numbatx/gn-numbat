package heartbeat_test

import (
	"bytes"
	"testing"

	"github.com/numbatx/gn-numbat/crypto"
	"github.com/numbatx/gn-numbat/node/heartbeat"
	"github.com/numbatx/gn-numbat/node/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewSender

func TestNewSender_NilP2pMessengerShouldErr(t *testing.T) {
	t.Parallel()

	sender, err := heartbeat.NewSender(
		nil,
		&mock.SinglesignStub{},
		&mock.PrivateKeyStub{},
		&mock.MarshalizerMock{},
		"",
	)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilMessenger, err)
}

func TestNewSender_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	sender, err := heartbeat.NewSender(
		&mock.MessengerStub{},
		nil,
		&mock.PrivateKeyStub{},
		&mock.MarshalizerMock{},
		"",
	)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilSingleSigner, err)
}

func TestNewSender_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	sender, err := heartbeat.NewSender(
		&mock.MessengerStub{},
		&mock.SinglesignStub{},
		nil,
		&mock.MarshalizerMock{},
		"",
	)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
}

func TestNewSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	sender, err := heartbeat.NewSender(
		&mock.MessengerStub{},
		&mock.SinglesignStub{},
		&mock.PrivateKeyStub{},
		nil,
		"",
	)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewSender_ShouldWork(t *testing.T) {
	t.Parallel()

	sender, err := heartbeat.NewSender(
		&mock.MessengerStub{},
		&mock.SinglesignStub{},
		&mock.PrivateKeyStub{},
		&mock.MarshalizerMock{},
		"",
	)

	assert.NotNil(t, sender)
	assert.Nil(t, err)
}

//------- SendHeartbeat

func TestSender_SendHeartbeatGeneratePublicKeyErrShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return nil, errExpected
		},
	}

	sender, _ := heartbeat.NewSender(
		&mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
			},
		},
		&mock.SinglesignStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
				return nil, nil
			},
		},
		&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return pubKey
			},
		},
		&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (i []byte, e error) {
				return nil, nil
			},
		},
		"",
	)

	err := sender.SendHeartbeat()

	assert.Equal(t, errExpected, err)
}

func TestSender_SendHeartbeatSignErrShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return nil, nil
		},
	}

	sender, _ := heartbeat.NewSender(
		&mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
			},
		},
		&mock.SinglesignStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
				return nil, errExpected
			},
		},
		&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return pubKey
			},
		},
		&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (i []byte, e error) {
				return nil, nil
			},
		},
		"",
	)

	err := sender.SendHeartbeat()

	assert.Equal(t, errExpected, err)
}

func TestSender_SendHeartbeatMarshalizerErrShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return nil, nil
		},
	}

	sender, _ := heartbeat.NewSender(
		&mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
			},
		},
		&mock.SinglesignStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
				return nil, nil
			},
		},
		&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				return pubKey
			},
		},
		&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (i []byte, e error) {
				return nil, errExpected
			},
		},
		"",
	)

	err := sender.SendHeartbeat()

	assert.Equal(t, errExpected, err)
}

func TestSender_SendHeartbeatShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	marshaledBuff := []byte("marshalBuff")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return []byte("pub key"), nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyClled := false
	marshalCalled := false

	sender, _ := heartbeat.NewSender(
		&mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				if topic == testTopic && bytes.Equal(buff, marshaledBuff) {
					broadcastCalled = true
				}
			},
		},
		&mock.SinglesignStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
				signCalled = true
				return signature, nil
			},
		},
		&mock.PrivateKeyStub{
			GeneratePublicHandler: func() crypto.PublicKey {
				genPubKeyClled = true
				return pubKey
			},
		},
		&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) (i []byte, e error) {
				hb, ok := obj.(*heartbeat.Heartbeat)
				if ok {
					pubkeyBytes, _ := pubKey.ToByteArray()
					if bytes.Equal(hb.Signature, signature) &&
						bytes.Equal(hb.Pubkey, pubkeyBytes) {

						marshalCalled = true
						return marshaledBuff, nil
					}
				}
				return nil, nil
			},
		},
		testTopic,
	)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyClled)
	assert.True(t, marshalCalled)
}
