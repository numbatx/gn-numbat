package mock

import (
	"github.com/numbatx/gn-numbat/p2p"
)

type MessengerStub struct {
	CloseCalled                    func() error
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	HasTopicCalled                 func(name string) bool
	HasTopicValidatorCalled        func(name string) bool
	BroadcastOnChannelCalled       func(channel string, topic string, buff []byte)
	BroadcastCalled                func(topic string, buff []byte)
	RegisterMessageProcessorCalled func(topic string, handler p2p.MessageProcessor) error
	BootstrapCalled                func() error
	PeerAddressCalled              func(pid p2p.PeerID) string
}

func (ms *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return ms.RegisterMessageProcessorCalled(topic, handler)
}

func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

func (ms *MessengerStub) Close() error {
	return ms.CloseCalled()
}

func (ms *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return ms.CreateTopicCalled(name, createChannelForTopic)
}

func (ms *MessengerStub) HasTopic(name string) bool {
	return ms.HasTopicCalled(name)
}

func (ms *MessengerStub) HasTopicValidator(name string) bool {
	return ms.HasTopicValidatorCalled(name)
}

func (ms *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	ms.BroadcastOnChannelCalled(channel, topic, buff)
}

func (ms *MessengerStub) Bootstrap() error {
	return ms.BootstrapCalled()
}

func (ms *MessengerStub) PeerAddress(pid p2p.PeerID) string {
	return ms.PeerAddressCalled(pid)
}
