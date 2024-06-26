package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
)

// PeerInfoHandler is the signature of the handler that gets called whenever an action for a peerInfo is triggered
type PeerInfoHandler func(pInfo peerstore.PeerInfo)

// ConnectableHost is an enhanced Host interface that has the ability to connect to a string address
type ConnectableHost interface {
	host.Host
	ConnectToPeer(ctx context.Context, address string) error
}

type connectableHost struct {
	host.Host
}

// NewConnectableHost creates a new connectable host implementation
func NewConnectableHost(h host.Host) *connectableHost {
	return &connectableHost{
		Host: h,
	}
}

// ConnectToPeer connects to a peer by knowing its string address
func (connHost *connectableHost) ConnectToPeer(ctx context.Context, address string) error {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	return connHost.Connect(ctx, *pInfo)
}
