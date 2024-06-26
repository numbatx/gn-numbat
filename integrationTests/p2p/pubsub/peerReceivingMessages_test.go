package peerDisconnecting

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/numbatx/gn-numbat/p2p"
	"github.com/numbatx/gn-numbat/p2p/libp2p"
	"github.com/numbatx/gn-numbat/p2p/libp2p/discovery"
	"github.com/numbatx/gn-numbat/p2p/loadBalancer"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var durationTest = time.Duration(time.Second * 30)
var randezVous = "numbatRandezVous"

type messageProcessorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

func (mps *messageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return mps.ProcessReceivedMessageCalled(message)
}

func createMessenger(ctx context.Context, seed int, initialPeerList []string) p2p.Messenger {

	r := rand.New(rand.NewSource(int64(seed)))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessengerOnFreePort(
		ctx,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewKadDhtPeerDiscoverer(time.Second, randezVous, initialPeerList))

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func TestPeerReceivesTheSameMessageMultipleTimesShouldNotHappen(t *testing.T) {

	if testing.Short() {
		t.Skip("this is not a short test")
	}

	noOfPeers := 20

	//Step 1. Create advertiser
	advertiser := createMessenger(context.Background(), noOfPeers, nil)

	//Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, noOfPeers)
	for i := 0; i < noOfPeers; i++ {
		node := createMessenger(context.Background(), i, []string{chooseNonCircuitAddress(advertiser.Addresses())})
		peers[i] = node
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < noOfPeers; i++ {
			if peers[i] != nil {
				peers[i].Close()
			}
		}

		if advertiser != nil {
			advertiser.Close()
		}
	}()

	chanStop := make(chan struct{})

	//Step 3. Register pubsub validators
	mutMapMessages := sync.Mutex{}
	mapMessages := make(map[int]map[string]struct{})
	testTopic := "test"

	for i := 0; i < noOfPeers; i++ {
		idx := i
		mapMessages[idx] = make(map[string]struct{})
		peers[idx].CreateTopic(testTopic, true)

		peers[idx].RegisterMessageProcessor(testTopic, &messageProcessorStub{
			ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
				time.Sleep(time.Second)

				mutMapMessages.Lock()
				defer mutMapMessages.Unlock()

				msgId := "peer: " + message.Peer().Pretty() + " - seqNo: 0x" + hex.EncodeToString(message.SeqNo())
				_, ok := mapMessages[idx][msgId]
				if ok {
					assert.Fail(t, "message %s received twice", msgId)
					chanStop <- struct{}{}
				}

				mapMessages[idx][msgId] = struct{}{}
				return nil
			},
		})
	}
	fmt.Println("step3 completed------->")

	//Step 4. Call bootstrap on all peers
	advertiser.Bootstrap()
	for _, p := range peers {
		p.Bootstrap()
	}
	waitForBootstrapAndShowConnected(peers)
	fmt.Println("step4 completed------->")

	//Step 5. Continuously send messages from one peer
	for timeStart := time.Now(); timeStart.Add(durationTest).Unix() > time.Now().Unix(); {
		peers[0].Broadcast(testTopic, []byte("test buff"))
		select {
		case <-chanStop:
			fmt.Println("stpe55555")
			return
		default:
			return
		}
		time.Sleep(time.Millisecond)
	}
}

func waitForBootstrapAndShowConnected(peers []p2p.Messenger) {
	fmt.Printf("Waiting %v for peer discovery...\n", durationBootstrapingTime)
	time.Sleep(durationBootstrapingTime)

	fmt.Println("Connected peers:")
	for _, p := range peers {
		fmt.Printf("Peer %s is connected to %d peers\n", p.ID().Pretty(), len(p.ConnectedPeers()))
	}
}

func chooseNonCircuitAddress(addresses []string) string {
	for _, adr := range addresses {
		if strings.Contains(adr, "circuit") || strings.Contains(adr, "169.254") {
			continue
		}
		return adr
	}

	return ""
}
