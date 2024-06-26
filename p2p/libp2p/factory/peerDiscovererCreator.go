package factory

import (
	"time"

	"github.com/numbatx/gn-numbat/config"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/p2p/libp2p/discovery"
)

type peerDiscovererCreator struct {
	p2pConfig config.P2PConfig
}

// NewPeerDiscovererCreator creates a new instance of peer discovery factory
func NewPeerDiscovererCreator(pConfig config.P2PConfig) *peerDiscovererCreator {
	return &peerDiscovererCreator{
		p2pConfig: pConfig,
	}
}

// CreatePeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func (pdc *peerDiscovererCreator) CreatePeerDiscoverer() (p2p.PeerDiscoverer, error) {
	isMoreThanOneEnabled := pdc.p2pConfig.MdnsPeerDiscovery.Enabled && pdc.p2pConfig.KadDhtPeerDiscovery.Enabled

	if isMoreThanOneEnabled {
		return nil, p2p.ErrMoreThanOnePeerDiscoveryActive
	}

	if pdc.p2pConfig.KadDhtPeerDiscovery.Enabled {
		return pdc.createKadDhtPeerDiscoverer()
	}

	if pdc.p2pConfig.MdnsPeerDiscovery.Enabled {
		return pdc.createMdnsPeerDiscoverer()
	}

	return discovery.NewNullDiscoverer(), nil
}

func (pdc *peerDiscovererCreator) createKadDhtPeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec <= 0 {
		return nil, p2p.ErrNegativeOrZeroPeersRefreshInterval
	}

	return discovery.NewKadDhtPeerDiscoverer(
		time.Second*time.Duration(pdc.p2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		pdc.p2pConfig.KadDhtPeerDiscovery.RandezVous,
		pdc.p2pConfig.KadDhtPeerDiscovery.InitialPeerList,
	), nil
}

func (pdc *peerDiscovererCreator) createMdnsPeerDiscoverer() (p2p.PeerDiscoverer, error) {
	if pdc.p2pConfig.MdnsPeerDiscovery.RefreshIntervalInSec <= 0 {
		return nil, p2p.ErrNegativeOrZeroPeersRefreshInterval
	}

	return discovery.NewMdnsPeerDiscoverer(
		time.Second*time.Duration(pdc.p2pConfig.MdnsPeerDiscovery.RefreshIntervalInSec),
		pdc.p2pConfig.MdnsPeerDiscovery.ServiceTag,
	), nil
}
