package blocace

import (
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	tip    []byte
	cursor []byte
	db     *bolt.DB
	search *Search
}

// RegisterAccount persists the account to the storage
func (bc Blockchain) RegisterAccount(address []byte, account Account) error {
	result := account.Serialize()

	err := bc.db.Update(func(dbtx *bolt.Tx) error {
		aBucket, _ := dbtx.CreateBucketIfNotExists([]byte(accountsBucket))
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

	err := bc.db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(blocksBucket))
		lastHash = bBucket.Get([]byte("l"))
		lastHeight = bBucket.Get([]byte("b"))

		return nil
	})

	lastHeightInt, err := strconv.ParseInt(string(lastHeight), 10, 64)

	newBlock := NewBlock(txs, lastHash, uint64(lastHeightInt+1))
	bc.tip, err = newBlock.Persist(bc.db)

	start := time.Now().UnixNano()
	log.Debug("number of transactions in the block:" + strconv.FormatInt(int64(newBlock.TotalTransactions), 10))
	log.Debug("start indexing the block:" + strconv.FormatInt(start, 10))
	bc.search.IndexBlock(newBlock)
	end := time.Now().UnixNano()
	log.Debug("end indexing the block:" + strconv.FormatInt(end, 10) + ", duration:" + strconv.FormatInt((end-start)/1000000, 10) + "ms")

	if err != nil {
		log.Error(err)
	}
}

// Next returns next block starting from the tip
func (bc *Blockchain) Next() *Block {
	var block *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(bc.cursor)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Error(err)
	}

	bc.cursor = block.PrevBlockHash

	return block
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new Blockchain with genesis Block (reading existing DB data and initializing a Blockchain struct)
func NewBlockchain(dbFile string) *Blockchain {
	if dbExists(dbFile) == false {
		log.Fatal("No existing blockchain found. Create one first.")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(blocksBucket))
		tip = bBucket.Get([]byte("l"))

		return nil
	})

	blockchainSearch, err := NewSearch(db, dataDir)

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, tip, db, blockchainSearch}

	return &bc
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(dbFile string) *Blockchain {
	if dbExists(dbFile) {
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

	bc := Blockchain{tip, tip, db, blockchainSearch}

	return &bc
}
