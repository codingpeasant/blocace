package p2p

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/codingpeasant/blocace/blockchain"

	log "github.com/sirupsen/logrus"
)

const peerBlockchainDir = "peers"

// BlockchainForest defines the local and peer chains
type BlockchainForest struct {
	local *blockchain.Blockchain
	peers map[string]*blockchain.Blockchain
}

func NewBlockchainForest(bcLocal *blockchain.Blockchain) *BlockchainForest {
	peers := make(map[string]*blockchain.Blockchain)
	peerBlockchainsDirRoot := bcLocal.DataDir + filepath.Dir("/") + peerBlockchainDir

	if blockchain.DbExists(peerBlockchainsDirRoot) == false {
		log.Infof("did not find peer db dir %s, creating one...", peerBlockchainsDirRoot)
		err := os.MkdirAll(peerBlockchainsDirRoot, 0600)
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

			peers[fmt.Sprintf("%x", peerId)] = &blockchain.Blockchain{Tip: tip, Db: db, Search: bcLocal.Search, DataDir: bcLocal.DataDir}

		}
	}

	return &BlockchainForest{bcLocal, peers}
}
