package blockchain

import (
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// Blockchain keeps a sequence of Blocks. Blockchain DB keys: lastHash - l; lastHeight - b; totalTransactions- t; p2pPrivKey;
type Blockchain struct {
	Tip     []byte
	Db      *bolt.DB
	Search  *Search
	DataDir string
}

// RegisterAccount persists the account to the storage
func (bc *Blockchain) RegisterAccount(address []byte, account Account) error {
	result := account.Marshal()

	err := bc.Db.Update(func(dbtx *bolt.Tx) error {
		aBucket, _ := dbtx.CreateBucketIfNotExists([]byte(AccountsBucket))
		err := aBucket.Put(address, result)
		if err != nil {
			log.Error(err)
		}

		return nil
	})

	return err
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(txs []*Transaction) {
	var lastHash []byte
	var lastHeight []byte

	err := bc.Db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(BlocksBucket))
		lastHash = bBucket.Get([]byte("l"))
		lastHeight = bBucket.Get([]byte("b"))

		return nil
	})

	lastHeightInt, err := strconv.ParseInt(string(lastHeight), 10, 64)

	newBlock := NewBlock(txs, lastHash, uint64(lastHeightInt+1))
	bc.Tip, err = newBlock.Persist(bc.Db)

	start := time.Now().UnixNano()
	log.Debug("start indexing the block:" + strconv.FormatInt(start, 10))
	bc.Search.IndexBlock(newBlock)
	end := time.Now().UnixNano()
	log.Debug("end indexing the block:" + strconv.FormatInt(end, 10) + ", duration:" + strconv.FormatInt((end-start)/1000000, 10) + "ms")

	if err != nil {
		log.Error(err)
	}
}

func DbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new Blockchain with genesis Block (reading existing DB data and initializing a Blockchain struct)
func NewBlockchain(dbFile string, dataDir string) *Blockchain {
	if DbExists(dbFile) == false {
		log.Fatal("No existing blockchain found. Create one first.")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(BlocksBucket))
		tip = bBucket.Get([]byte("l"))

		return nil
	})

	blockchainSearch, err := NewSearch(db, dataDir)

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db, blockchainSearch, dataDir}

	return &bc
}

// CreateBlockchain creates a new local blockchain DB
func CreateBlockchain(dbFile string, dataDir string) *Blockchain {
	if DbExists(dbFile) {
		log.Fatal("Blockchain already exists.")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	cbtx := NewCoinbaseTX()
	genesisBlock := NewGenesisBlock(cbtx, db)

	tip, err = genesisBlock.Persist(db)

	if err != nil {
		log.Panic(err)
	}

	blockchainSearch, err := NewSearch(db, dataDir)

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db, blockchainSearch, dataDir}

	return &bc
}
