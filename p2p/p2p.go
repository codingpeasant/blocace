package p2p

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/patrickmn/go-cache"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"

	"github.com/codingpeasant/blocace/blockchain"
)

var DefaultPort = 6091

// P2P is the main object to handle networking-related messages
type P2P struct {
	Node                *noise.Node
	BlockchainForest    *BlockchainForest
	ChallengeWordsCache *cache.Cache
	overlay             *kademlia.Protocol
	Accounts            map[string]blockchain.Account // used by http
	mappings            map[string]blockchain.DocumentMapping
}

// BroadcastObject sends a serializable object to all the known peers
func (p *P2P) BroadcastObject(object noise.Serializable) {
	// add the account(s) to local cache before broadcasting
	accountsToAdd, ok := object.(AccountsP2P)
	if ok {
		for address, account := range accountsToAdd.Accounts {
			p.Accounts[address] = account // update the cache
		}
	}

	// add the mapping(s) to local cache before broadcasting
	mappingsToAdd, ok := object.(MappingsP2P)
	if ok {
		for collection, mapping := range mappingsToAdd.Mappings {
			p.mappings[collection] = mapping // update the cache
		}
	}

	for _, id := range p.overlay.Table().Peers() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

//
func (p *P2P) GetPeers() []byte {
	var peers []noise.ID
	for _, peer := range p.overlay.Table().Peers() {
		peers = append(peers, peer)
	}
	peersJSON, _ := json.Marshal(peers)

	return peersJSON
}

// SyncAccountsFromPeers sends rpc to peers to sync the accounts
func (p *P2P) SyncAccountsFromPeers() {
	for _, id := range p.overlay.Table().Peers() {
		sendAccountsRequest(p.Accounts, p.Node, id, p.BlockchainForest.Local)
	}
}

// SyncMappingsFromPeers sends rpc to peers to sync the mappings
func (p *P2P) SyncMappingsFromPeers() {
	for _, id := range p.overlay.Table().Peers() {
		sendMappingsRequest(p.mappings, p.Node, id, p.BlockchainForest.Local.Search)
	}
}

// SyncPeerBlockchains sends rpc to all known peers to sync the peer blockchains to local
func (p *P2P) SyncPeerBlockchains() {
	for _, id := range p.overlay.Table().Peers() {
		syncPeerBlockchain(p.Node, id, p.BlockchainForest, false) // startup sync, not reverse sync
	}
}

// NewP2P initializes the P2P node with messages and handlers
func NewP2P(bc *blockchain.Blockchain, bindHost string, bindPort uint16, advertiseAddress string, connectionAddresses ...string) *P2P {
	accounts := initializeAccounts(bc.Db)
	mappings := initializeMappings(bc.Db)

	blockchainForest := NewBlockchainForest(bc)
	challengeWordsCache := cache.New(30*time.Second, 1*time.Minute)

	var p2pPrivKey noise.PrivateKey
	var p2pPrivKeyBytes []byte

	// make sure to reuse the priv key
	err := bc.Db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(blockchain.BlocksBucket))
		p2pPrivKeyBytes = bBucket.Get([]byte(blockchain.P2PPrivateKeyKey))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	copy(p2pPrivKey[:], p2pPrivKeyBytes)

	// Create a new configured node.
	node, err := noise.NewNode(
		noise.WithNodeBindHost(net.ParseIP(bindHost)),
		noise.WithNodeBindPort(bindPort),
		noise.WithNodeAddress(advertiseAddress),
		noise.WithNodePrivateKey(p2pPrivKey),
	)

	if err != nil {
		log.Panic(err)
	}

	// Register the chatMessage Go type to the node with an associated unmarshal function.
	node.RegisterMessage(RequestP2P{}, unmarshalRequestP2P)
	node.RegisterMessage(AccountsP2P{}, unmarshalAccountsP2P)
	node.RegisterMessage(MappingsP2P{}, unmarshalMappingsP2P)
	node.RegisterMessage(ChallengeWordP2P{}, unmarshalChallengeWordP2P)
	node.RegisterMessage(BlockP2P{}, unmarshalBlockP2P)

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

			switch requestP2P.RequestType {
			case accountsRequestType:
				ctx.SendMessage(handleAccountsRequest(requestP2P, accounts))
				accountsRequestReverse(requestP2P, accounts, node, ctx.ID(), blockchainForest.Local) // sync new accounts from remote
			case mappingsRequestType:
				ctx.SendMessage(handleMappingsRequest(requestP2P, mappings))
				mappingsRequestReverse(requestP2P, mappings, node, ctx.ID(), blockchainForest.Local.Search) // sync new mappings from remote
			case blockRequestType:
				ctx.SendMessage(handleBlockRequest(requestP2P, blockchainForest))
				// reversely sync the blockchain but don't reverse the reverse
				if requestP2P.RequestParameters["local"] == "tip" && requestP2P.RequestParameters["reverse"] != "reverse" {
					syncPeerBlockchain(node, ctx.ID(), blockchainForest, true)
				}
			default:
				log.Warnf("got unsupported RequestP2P request type: +%v", requestP2P)
				return nil
			}

		} else {
			obj, err := ctx.DecodeMessage()
			if err != nil {
				return err
			}

			switch objectP2p := obj.(type) {
			case AccountsP2P:
				for address, account := range objectP2p.Accounts {
					if funk.IsEmpty(accounts[address]) || accounts[address].LastModified < account.LastModified {
						accounts[address] = account // update the cache
						if err = bc.RegisterAccount([]byte(address), account); err != nil {
							return err
						}
					}
				}
			case MappingsP2P:
				for mappingName, mapping := range objectP2p.Mappings {
					if funk.IsEmpty(mappings[mappingName]) {
						mappings[mappingName] = mapping // update the cache
						if _, err = bc.Search.CreateMapping(mapping); err != nil {
							return err
						}
					}
				}
			case ChallengeWordP2P:
				if funk.IsEmpty(objectP2p.Address) && funk.IsEmpty(objectP2p.ChallengeWord) {
					challengeWordsCache.Set(objectP2p.ChallengeWord, objectP2p.Address, cache.DefaultExpiration)
				}
			case BlockP2P:
				log.Debugf("BlockFromPeer: %s(%s) > %+x; height: %d\n", ctx.ID().Address, ctx.ID().ID.String(), objectP2p.Hash, objectP2p.Height)
				blockchainForest.AddBlock(objectP2p)
			case RequestP2P:
				ctx.SendMessage(handleBlockRequest(objectP2p, blockchainForest))
			default:
				return errors.New("cannot parse the object from peer: " + ctx.ID().ID.String())
			}
		}

		return nil
	})

	// Instantiate Kademlia.
	events := kademlia.Events{
		OnPeerAdmitted: func(id noise.ID) {
			log.Infof("learned about a new peer %s (%s).\n", id.Address, id.ID.String())
		},
		OnPeerEvicted: func(id noise.ID) {
			log.Infof("forgotten a peer %s (%s).\n", id.Address, id.ID.String())
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

	return &P2P{Node: node, overlay: overlay, BlockchainForest: blockchainForest, ChallengeWordsCache: challengeWordsCache, Accounts: accounts, mappings: mappings}
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
		str = append(str, fmt.Sprintf("%s (%s)", id.Address, id.ID.String()))
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
			accountMap[fmt.Sprintf("%s", k)] = *blockchain.DeserializeAccount(v)
		}

		return nil
	})
	return accountMap
}

// initializeMappings loads the mappings (schemas) from disk to RAM
func initializeMappings(db *bolt.DB) map[string]blockchain.DocumentMapping {
	collectionMapings := make(map[string]blockchain.DocumentMapping)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(blockchain.CollectionsBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			collectionMapings[fmt.Sprintf("%s", k)] = *blockchain.DeserializeDocumentMapping(v)
		}

		return nil
	})
	return collectionMapings
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
			if err != nil {
				continue
			}
			if account.LastModified > peerLastModified {
				accountsToSend[address] = account
			}
		}
	}

	return AccountsP2P{Accounts: accountsToSend}
}

// handleMappingsRequest returns the new mappings which the peer doesn't have
func handleMappingsRequest(request RequestP2P, mappingsLocal map[string]blockchain.DocumentMapping) MappingsP2P {
	mappingsToSend := make(map[string]blockchain.DocumentMapping)
	for collectionName, mapping := range mappingsLocal {
		if collectionName == "default" {
			continue
		}
		if funk.IsEmpty(request.RequestParameters[collectionName]) {
			mappingsToSend[collectionName] = mapping
		}
	}

	return MappingsP2P{Mappings: mappingsToSend}
}

// handleBlockRequest handles 1) local tip block; 2) local block; 3) peer block
func handleBlockRequest(request RequestP2P, bf *BlockchainForest) BlockP2P {
	var blockToReturn BlockP2P
	if !funk.IsEmpty(request.RequestParameters["local"]) {
		requestValue := request.RequestParameters["local"]
		if requestValue == "tip" {
			blockToReturn = bf.GetBlock(bf.Local.PeerId, bf.Local.Tip, false)
			blockToReturn.IsTip = true // mark as tip for peer to process

		} else {
			blockId, err := hex.DecodeString(requestValue)
			if err != nil {
				log.Error(err)
				return blockToReturn
			}
			blockToReturn = bf.GetBlock(bf.Local.PeerId, blockId, false)
		}

	} else {
		for peerIdString, blockIdString := range request.RequestParameters { // should be only one round

			blockId, err := hex.DecodeString(blockIdString)
			peerId, err := hex.DecodeString(peerIdString)
			if err != nil {
				log.Error(err)
				return blockToReturn
			}
			blockToReturn = bf.GetBlock(peerId, blockId, false)
		}
	}

	return blockToReturn
}

// mappingsRequestReverse checks if a peer has mapping(s) that is new and request for them
func mappingsRequestReverse(request RequestP2P, mappingsLocal map[string]blockchain.DocumentMapping, node *noise.Node, id noise.ID, search *blockchain.Search) {
	for _, mapping := range request.RequestParameters {
		if mapping == "default" {
			continue
		}

		if funk.IsEmpty(mappingsLocal[mapping]) {
			sendMappingsRequest(mappingsLocal, node, id, search)
			break
		}
	}
}

// accountsRequestReverse checks if a peer has accounts(s) that is new and request for them
func accountsRequestReverse(request RequestP2P, accountsLocal map[string]blockchain.Account, node *noise.Node, id noise.ID, bcLocal *blockchain.Blockchain) {
	for address, peerLastModified := range request.RequestParameters {
		if funk.IsEmpty(accountsLocal[address]) {
			sendAccountsRequest(accountsLocal, node, id, bcLocal)
			break
		} else {
			peerLastModifiedLong, err := strconv.ParseInt(peerLastModified, 10, 64)
			if err != nil {
				continue
			}
			if peerLastModifiedLong > accountsLocal[address].LastModified {
				sendAccountsRequest(accountsLocal, node, id, bcLocal)
				break
			}
		}
	}
}

// syncPeerBlockchain sends rpc to a peer to sync the peer blockchain to local
func syncPeerBlockchain(node *noise.Node, id noise.ID, bf *BlockchainForest, reverse bool) {
	log.Infof("start syncing blocks from peer %s (%s)...", id.Address, id.ID.String())
	// first update the tip
	requestParameters := make(map[string]string)
	requestParameters["local"] = "tip"
	if reverse {
		requestParameters["reverse"] = "reverse" // distinguish between startup sync and reverse sync
	}

	previousBlockHash := sendBlockRequest(requestParameters, node, id, bf)

	// check blocks
	var previousBlock BlockP2P
	// until genesis block is reached or cannot find previous
	for bytes.Compare(previousBlockHash, []byte{}) != 0 {
		previousBlock = bf.GetBlock(id.ID[:], previousBlockHash, true)
		if !funk.IsEmpty(previousBlock) {
			previousBlockHash = previousBlock.PrevBlockHash
		} else {
			requestParameters["local"] = fmt.Sprintf("%x", previousBlockHash)
			previousBlockHash = sendBlockRequest(requestParameters, node, id, bf)
		}
	}
	log.Infof("finished syncing from peer %s (%s)", id.Address, id.ID.String())
}

// sendMappingsRequest and update local mapping cache
func sendMappingsRequest(mappingsLocal map[string]blockchain.DocumentMapping, node *noise.Node, id noise.ID, search *blockchain.Search) {
	requestParameters := make(map[string]string)
	for collectionName := range mappingsLocal {
		requestParameters[collectionName] = collectionName // collectionName:collectionName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mappingsFromPeerRes, err := node.RequestMessage(ctx, id.Address, RequestP2P{RequestType: mappingsRequestType, RequestParameters: requestParameters})
	cancel()

	if err != nil {
		log.Errorf("failed to send mapping request message to %s(%s). Skipping... [error: %s]\n",
			id.Address,
			id.ID.String(),
			err,
		)
		return
	}

	mappingsFromPeer, ok := mappingsFromPeerRes.(MappingsP2P)
	if !ok {
		log.Error("cannot parse mappings from peer: " + id.ID.String())
	}

	for collectionName, mapping := range mappingsFromPeer.Mappings {
		if funk.IsEmpty(mappingsLocal[collectionName]) {
			mappingsLocal[collectionName] = mapping // update the cache
			log.Debugf("Collection: %s(%s) > %+v\n", id.Address, id.ID.String(), mapping)
			if _, err = search.CreateMapping(mapping); err != nil {
				log.Error(err)
			}
		}
	}

}

// sendAccountsRequest and update local accounts cache
func sendAccountsRequest(accountsLocal map[string]blockchain.Account, node *noise.Node, id noise.ID, bcLocal *blockchain.Blockchain) {
	requestParameters := make(map[string]string)
	for address, account := range accountsLocal {
		if account.Role.Name != "admin" {
			requestParameters[address] = strconv.Itoa(int(account.LastModified)) // address:lastModified
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	accountsFromPeerRes, err := node.RequestMessage(ctx, id.Address, RequestP2P{RequestType: accountsRequestType, RequestParameters: requestParameters})
	cancel()

	if err != nil {
		log.Errorf("failed to send account request message to %s(%s). Skipping... [error: %s]\n",
			id.Address,
			id.ID.String(),
			err,
		)
	}

	accountsFromPeer, ok := accountsFromPeerRes.(AccountsP2P)
	if !ok {
		log.Error("cannot parse accounts from peer: " + id.ID.String())
	}

	for address, account := range accountsFromPeer.Accounts {
		if funk.IsEmpty(accountsLocal[address]) || accountsLocal[address].LastModified < account.LastModified {
			accountsLocal[address] = account // update the cache
			log.Debugf("Account: %s(%s) > %+v\n", id.Address, id.ID.String(), account)
			if err = bcLocal.RegisterAccount([]byte(address), account); err != nil {
				log.Error(err)
			}
		}
	}
}

// sendBlockRequest and update peer blockchain locally
func sendBlockRequest(requestParameters map[string]string, node *noise.Node, id noise.ID, bf *BlockchainForest) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	blockFromPeerRes, err := node.RequestMessage(ctx, id.Address, RequestP2P{RequestType: blockRequestType, RequestParameters: requestParameters})
	cancel()

	if err != nil {
		log.Warnf("failed to send block request message to %s(%s). retrying... [error: %s]\n",
			id.Address,
			id.ID.String(),
			err,
		)
		return sendBlockRequest(requestParameters, node, id, bf)
	}

	blockFromPeer, ok := blockFromPeerRes.(BlockP2P)
	if !ok {
		log.Error("cannot parse block from peer: " + id.ID.String())
	}
	log.Debugf("BlockFromPeer: %s(%s) > %+x; height: %d\n", id.Address, id.ID.String(), blockFromPeer.Hash, blockFromPeer.Height)

	bf.AddBlock(blockFromPeer)
	return blockFromPeer.PrevBlockHash
}
