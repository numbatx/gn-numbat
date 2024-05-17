package topicResolverSender

import (
	"github.com/numbatx/gn-numbat/dataRetriever"
	"github.com/numbatx/gn-numbat/p2p"
)

func SelectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer dataRetriever.IntRandomizer) ([]p2p.PeerID, error) {
	return selectRandomPeers(connectedPeers, peersToSend, randomizer)
}
