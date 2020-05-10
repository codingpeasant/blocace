package p2p

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/codingpeasant/blocace/blockchain"

	log "github.com/sirupsen/logrus"
)

const peerBlockchainDir = "peers"

// BlockchainForest defines the local and peer chains
type BlockchainForest struct {
	Local *blockchain.Blockchain
	Peers map[string]*blockchain.Blockchain
}

// AddBlockBroadcasted persist the broadcasted new block from a peer to local peer blockchain db and index it
func (b *BlockchainForest) AddBlockBroadcasted(blockP2p BlockP2P) {
	peerIdStr := fmt.Sprintf("%x", blockP2p.PeerId)

	block, err := blockP2p.MapToBlock()

	if err != nil {
		log.Error(err)
		return
	}

	if bytes.Compare(block.Hash, block.SetHash()) != 0 {
		log.Errorf("block hash verification failed, abandon this block")
		return
	}

	if b.Peers[peerIdStr] == nil {
		log.Infof("peer %s blockchain db not found, creating one...", peerIdStr)
		peerBlockchainsDbFile := b.Local.DataDir + filepath.Dir("/") + peerBlockchainDir + filepath.Dir("/") + fmt.Sprintf("%x", blockP2p.PeerId) + ".db"

		db, err := bolt.Open(peerBlockchainsDbFile, 0600, nil)
		if err != nil {
			log.Panic(err)
		}

		err = db.Update(func(dbtx *bolt.Tx) error {
			bBucket, err := dbtx.CreateBucket([]byte(blockchain.BlocksBucket))
			if err != nil {
				log.Panic(err)
			}

			_, err = dbtx.CreateBucket([]byte(blockchain.TransactionsBucket))
			if err != nil {
				log.Panic(err)
			}

			err = bBucket.Put([]byte("peerId"), blockP2p.PeerId)
			if err != nil {
				return err
			}

			return nil
		})

		b.Peers[peerIdStr] = &blockchain.Blockchain{Tip: blockP2p.Hash, PeerId: blockP2p.PeerId, Db: db, Search: b.Local.Search, DataDir: b.Local.DataDir}

	}

	_, err = block.Persist(b.Peers[peerIdStr].Db)
	if err != nil {
		log.Error(err)
	}

	start := time.Now().UnixNano()
	log.Debugf("start indexing the block at %d for peer blockchain %s...", start, peerIdStr)
	b.Local.Search.IndexBlock(block, blockP2p.PeerId)
	end := time.Now().UnixNano()
	log.Debug("end indexing the block:" + strconv.FormatInt(end, 10) + ", duration:" + strconv.FormatInt((end-start)/1000000, 10) + "ms")
}

// NewBlockchainForest initializes the peer blockchains by reading existing dbs from peerBlockchainDir which will be created should not exist
func NewBlockchainForest(bcLocal *blockchain.Blockchain) *BlockchainForest {
	peers := make(map[string]*blockchain.Blockchain)
	peerBlockchainsDirRoot := bcLocal.DataDir + filepath.Dir("/") + peerBlockchainDir

	if blockchain.DbExists(peerBlockchainsDirRoot) == false {
		log.Infof("did not find peer db dir %s, creating one...", peerBlockchainsDirRoot)
		err := os.MkdirAll(peerBlockchainsDirRoot, 0700)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Info("opening existing peer blockchains...")
		files, err := ioutil.ReadDir(peerBlockchainsDirRoot)
		if err != nil {
			log.Fatal(err)
		}

		// initialize all other peer blockchains than the local
		for _, file := range files {
			db, err := bolt.Open(peerBlockchainsDirRoot+filepath.Dir("/")+file.Name(), 0600, nil)
			if err != nil {
				log.Warnf("cannot open blockchain db %s: %s", peerBlockchainsDirRoot+filepath.Dir("/")+file.Name(), err.Error())
				continue
			}

			var tip, peerId []byte
			err = db.View(func(dbtx *bolt.Tx) error {
				bBucket := dbtx.Bucket([]byte(blockchain.BlocksBucket))
				tip = bBucket.Get([]byte("l"))

				return nil
			})

			if err != nil {
				log.Warnf("cannot get tip of blockchain at %s: %s", peerBlockchainsDirRoot+filepath.Dir("/")+file.Name(), err.Error())
				continue
			}

			err = db.View(func(dbtx *bolt.Tx) error {
				bBucket := dbtx.Bucket([]byte(blockchain.BlocksBucket))
				peerId = bBucket.Get([]byte("peerId"))

				return nil
			})

			if err != nil {
				log.Warnf("cannot get peerId of blockchain at %s: %s", peerBlockchainsDirRoot+filepath.Dir("/")+file.Name(), err.Error())
				continue
			}
			peers[fmt.Sprintf("%x", peerId)] = &blockchain.Blockchain{Tip: tip, PeerId: peerId, Db: db, DataDir: bcLocal.DataDir}

		}
	}

	return &BlockchainForest{bcLocal, peers}
}
