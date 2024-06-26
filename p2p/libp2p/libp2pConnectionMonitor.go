package libp2p

import (
	"time"

	net "github.com/libp2p/go-libp2p-net"
	"github.com/multiformats/go-multiaddr"
	"github.com/numbatx/gn-numbat/p2p"
)

// ThresholdMinimumConnectedPeers if the number of connected peers drop under this value, for each disconnecting
// peer, a trigger to reconnect to initial peers is done
var ThresholdMinimumConnectedPeers = 3

// DurationBetweenReconnectAttempts is used as to not call reconnecter.ReconnectToNetwork() to often
// when there are a lot of peers disconnecting and reconnection to initial nodes succeed
var DurationBetweenReconnectAttempts = time.Duration(time.Second * 5)

type libp2pConnectionMonitor struct {
	chDoReconnect chan struct{}
	reconnecter   p2p.Reconnecter
}

func newLibp2pConnectionMonitor(reconnecter p2p.Reconnecter) *libp2pConnectionMonitor {
	cm := &libp2pConnectionMonitor{
		reconnecter:   reconnecter,
		chDoReconnect: make(chan struct{}, 0),
	}

	if reconnecter != nil {
		go cm.doReconnection()
	}

	return cm
}

// Listen is called when network starts listening on an addr
func (lcm *libp2pConnectionMonitor) Listen(net.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcm *libp2pConnectionMonitor) ListenClose(net.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (lcm *libp2pConnectionMonitor) Connected(net.Network, net.Conn) {}

// Disconnected is called when a connection closed
func (lcm *libp2pConnectionMonitor) Disconnected(netw net.Network, conn net.Conn) {
	if len(netw.Conns()) < ThresholdMinimumConnectedPeers {
		select {
		case lcm.chDoReconnect <- struct{}{}:
		default:
		}
	}
}

// OpenedStream is called when a stream opened
func (lcm *libp2pConnectionMonitor) OpenedStream(net.Network, net.Stream) {}

// ClosedStream is called when a stream closed
func (lcm *libp2pConnectionMonitor) ClosedStream(net.Network, net.Stream) {}

func (lcm *libp2pConnectionMonitor) doReconnection() {
	for {
		select {
		case <-lcm.chDoReconnect:
			<-lcm.reconnecter.ReconnectToNetwork()
		}

		time.Sleep(DurationBetweenReconnectAttempts)
	}
}
