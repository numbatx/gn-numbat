package mock

import (
	"github.com/numbatx/gn-numbat/p2p"
)

type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P) error
}

func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return mps.ProcessMessageCalled(message)
}
