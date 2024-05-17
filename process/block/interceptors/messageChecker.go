package interceptors

import (
	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/process"
)

var log = logger.DefaultLogger()

type messageChecker struct {
}

func (*messageChecker) checkMessage(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	return nil
}
