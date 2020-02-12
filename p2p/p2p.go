package p2p

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

var DefaultPort = 6091

type P2P struct {
	node *noise.Node
}

func NewP2P(bindHost string, bindPort uint16, advertiseAddress string, connectionAddress ...string) *P2P {
	var nodeOptions []noise.NodeOption
	if !funk.IsEmpty(bindHost) {
		nodeOptions = append(nodeOptions, noise.WithNodeBindHost(net.ParseIP(bindHost)))
	}

	if !funk.IsEmpty(bindPort) {
		nodeOptions = append(nodeOptions, noise.WithNodeBindPort(bindPort))
	} else {
		log.Infof("no listening port provided. Using the default %d", DefaultPort)
		nodeOptions = append(nodeOptions, noise.WithNodeBindPort(uint16(DefaultPort)))
	}

	if !funk.IsEmpty(advertiseAddress) {
		nodeOptions = append(nodeOptions, noise.WithNodeAddress(advertiseAddress))
	}

	nodeOptions = append(nodeOptions, noise.WithNodeBindHost(net.ParseIP(bindHost)))

	if !funk.IsEmpty(advertiseAddress) {
		nodeOptions = append(nodeOptions, noise.WithNodeBindHost(net.ParseIP(bindHost)))
	}
	// Create a new configured node.
	node, err := noise.NewNode(nodeOptions...)

	if err != nil {
		log.Panic(err)
	}

	// Release resources associated to node at the end of the program.
	defer node.Close()

	// Instantiate Kademlia.
	events := kademlia.Events{
		OnPeerAdmitted: func(id noise.ID) {
			log.Infof("Learned about a new peer %s(%s).\n", id.Address, id.ID.String())
		},
		OnPeerEvicted: func(id noise.ID) {
			log.Infof("Forgotten a peer %s(%s).\n", id.Address, id.ID.String())
		},
		OnPeerActivity: func(id noise.ID) {
			log.Infof("OnPeerActivity: %s(%s).\n", id.Address, id.ID.String())
		},
	}

	overlay := kademlia.New(kademlia.WithProtocolEvents(events))

	// Bind Kademlia to the node.
	node.Bind(overlay.Protocol())

	// Have the node start listening for new peers.
	err = node.Listen()
	if err != nil {
		log.Panic(err)
	}

	if !funk.IsEmpty(connectionAddress) {
		// Ping nodes to initially bootstrap and discover peers from.
		bootstrap(node, connectionAddress)
		// Attempt to discover peers if we are bootstrapped to any nodes.
		discover(overlay)
	} else {
		log.Info("no peer address(es) provided, starting without trying to discover")
	}

	return &P2P{node: node}
}

// bootstrap pings and dials an array of network addresses which we may interact with and  discover peers from.
func bootstrap(node *noise.Node, addresses []string) {
	for _, addr := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := node.Ping(ctx, addr)
		cancel()

		if err != nil {
			log.Warnf("failed to ping bootstrap node (%s). [%s]\n", addr, err)
			continue
		}
	}
}

// discover uses Kademlia to discover new peers from nodes we already are aware of.
func discover(overlay *kademlia.Protocol) {
	ids := overlay.Discover()

	var str []string
	for _, id := range ids {
		str = append(str, fmt.Sprintf("%s(%s)", id.Address, id.ID.String()))
	}

	if len(ids) > 0 {
		log.Infof("discovered %d peer(s): [%v]\n", len(ids), strings.Join(str, ", "))
	} else {
		log.Warn("did not discover any peers.")
	}
}
