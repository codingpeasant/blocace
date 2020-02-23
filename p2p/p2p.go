package p2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"

	"github.com/codingpeasant/blocace/blockchain"
)

var DefaultPort = 6091

// P2P is the main object to handle networking-related messages
type P2P struct {
	Node     *noise.Node
	overlay  *kademlia.Protocol
	bc       *blockchain.Blockchain
	accounts map[string]blockchain.Account
}

// BroadcastObject sends a serializable object to all the known peers
func (p *P2P) BroadcastObject(object noise.Serializable) {
	for _, id := range p.overlay.Table().Peers() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := p.Node.SendMessage(ctx, id.Address, object)
		cancel()

		if err != nil {
			log.Errorf("failed to send object to %s(%s). Skipping... [error: %s]\n",
				id.Address,
				id.ID.String(),
				err,
			)
			continue
		}
	}
}

// BroadcastObject sends a serializable object to all the known peers
func (p *P2P) SyncAccountsFromPeers() {
	requestParameters := make(map[string]string)
	for address, account := range p.accounts {
		requestParameters[address] = strconv.Itoa(int(account.LastModified))
	}

	for _, id := range p.overlay.Table().Peers() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		accountsFromPeerRes, err := p.Node.RequestMessage(ctx, id.Address, RequestP2P{RequestType: "accounts", RequestParameters: requestParameters})
		cancel()

		if err != nil {
			log.Errorf("failed to send request message to %s(%s). Skipping... [error: %s]\n",
				id.Address,
				id.ID.String(),
				err,
			)
			continue
		}

		accountsFromPeer, ok := accountsFromPeerRes.(AccountsP2P)
		if !ok {
			log.Error("cannot parse account from peer: " + id.ID.String())
		}

		for address, account := range accountsFromPeer.Accounts {
			if funk.IsEmpty(p.accounts[address]) || p.accounts[address].LastModified < account.LastModified {
				p.accounts[address] = account // update the cache
				log.Debugf("%s(%s) > %+v\n", id.Address, id.ID.String(), account)
				if err = p.bc.RegisterAccount([]byte(address), account); err != nil {
					log.Error(err)
				}
			}
		}
	}
}

func NewP2P(bc *blockchain.Blockchain, bindHost string, bindPort uint16, advertiseAddress string, connectionAddresses ...string) *P2P {
	accounts := initializeAccounts(bc.Db)
	// Create a new configured node.
	node, err := noise.NewNode(
		noise.WithNodeBindHost(net.ParseIP(bindHost)),
		noise.WithNodeBindPort(bindPort),
		noise.WithNodeAddress(advertiseAddress),
	)

	if err != nil {
		log.Panic(err)
	}

	// Register the chatMessage Go type to the node with an associated unmarshal function.
	node.RegisterMessage(AccountsP2P{}, unmarshalAccountsP2P)
	node.RegisterMessage(RequestP2P{}, unmarshalRequestP2P)

	// Register a message handler to the node.
	node.Handle(func(ctx noise.HandlerContext) error {
		if ctx.IsRequest() {
			obj, err := ctx.DecodeMessage()
			if err != nil {
				return err
			}
			requestP2P, ok := obj.(RequestP2P)
			if !ok {
				return nil
			}
			return ctx.SendMessage(handleAccountsRequest(requestP2P, accounts))
		}

		obj, err := ctx.DecodeMessage()
		if err != nil {
			return err
		}

		accountsP2p, ok := obj.(AccountsP2P)
		if !ok {
			return errors.New("cannot parse account from peer: " + ctx.ID().ID.String())
		}

		for address, account := range accountsP2p.Accounts {
			if funk.IsEmpty(accounts[address]) || accounts[address].LastModified < account.LastModified {
				accounts[address] = account // update the cache
				log.Debugf("%s(%s) > %+v\n", ctx.ID().Address, ctx.ID().ID.String(), accountsP2p)
				if err = bc.RegisterAccount([]byte(address), account); err != nil {
					return err
				}
			}
		}

		return nil
	})

	// Instantiate Kademlia.
	events := kademlia.Events{
		OnPeerAdmitted: func(id noise.ID) {
			log.Infof("Learned about a new peer %s(%s).\n", id.Address, id.ID.String())
		},
		OnPeerEvicted: func(id noise.ID) {
			log.Infof("Forgotten a peer %s(%s).\n", id.Address, id.ID.String())
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

	if !funk.IsEmpty(connectionAddresses) {
		// Ping nodes to initially bootstrap and discover peers from.
		bootstrap(node, connectionAddresses)
		// Attempt to discover peers if we are bootstrapped to any nodes.
		discover(overlay)
	} else {
		log.Info("no peer address(es) provided, starting without trying to discover")
	}

	return &P2P{Node: node, overlay: overlay, bc: bc, accounts: accounts}
}

// bootstrap pings and dials an array of network addresses which we may interact with and  discover peers from.
func bootstrap(node *noise.Node, addresses []string) {
	for _, addr := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// initializeAccounts loads the accounts from disk to RAM
func initializeAccounts(db *bolt.DB) map[string]blockchain.Account {
	accountMap := make(map[string]blockchain.Account)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(blockchain.AccountsBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			accountMap[fmt.Sprintf("%x", k)] = *blockchain.DeserializeAccount(v)
		}

		return nil
	})
	return accountMap
}

// handleAccountsRequest returns the new accounts which the peer doesn't have or has a older version of
func handleAccountsRequest(request RequestP2P, accountsLocal map[string]blockchain.Account) AccountsP2P {
	accountsToSend := make(map[string]blockchain.Account)
	for address, account := range accountsLocal {
		if account.Role.Name == "admin" {
			continue
		}
		if funk.IsEmpty(request.RequestParameters[address]) {
			accountsToSend[address] = account
		} else {
			peerLastModified, err := strconv.ParseInt(request.RequestParameters[address], 10, 64)
			if err == nil {
				continue
			}
			if account.LastModified > peerLastModified {
				accountsToSend[address] = account
			}
		}
	}

	return AccountsP2P{Accounts: accountsToSend}
}
