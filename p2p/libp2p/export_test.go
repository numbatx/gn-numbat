package libp2p

import (
	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

var MaxSendBuffSize = maxSendBuffSize

func (netMes *networkMessenger) ConnManager() ifconnmgr.ConnManager {
	return netMes.ctxProvider.connHost.ConnManager()
}

func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.ctxProvider.connHost = newHost
}

func (ds *directSender) ProcessReceivedDirectMessage(message *pubsub_pb.Message) error {
	return ds.processReceivedDirectMessage(message)
}

func (ds *directSender) SeenMessages() *timecache.TimeCache {
	return ds.seenMessages
}

func (ds *directSender) Counter() uint64 {
	return ds.counter
}
