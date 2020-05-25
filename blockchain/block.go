package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// Block keeps block headers
type Block struct {
	Timestamp         int64
	PrevBlockHash     []byte
	Height            uint64
	Hash              []byte
	TotalTransactions int
	Transactions      []*Transaction
}

// Serialize serializes the block
func (b *Block) serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	transactions := b.Transactions
	b.Transactions = nil // don't encode transactions
	err := encoder.Encode(b)
	if err != nil {
		log.Error(err)
	}

	b.Transactions = transactions

	return result.Bytes()
}

// SetHash set the hash of the whole block
func (b *Block) SetHash() []byte {
	blockHash := sha256.Sum256(bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			b.GetMerkleTree().RootNode.Data,
			IntToHex(b.Timestamp),
		},
		[]byte{},
	))

	return blockHash[:]
}

// GetMerkleTree builds a merkle tree of all the transactions in the block
func (b *Block) GetMerkleTree() *MerkleTree {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	return NewMerkleTree(txHashes)
}

// Persist stores the block with the transactions to DB
func (b Block) Persist(db *bolt.DB, isTip bool) ([]byte, error) {
	var currentTxTotal []byte
	var currentTxTotalInt int64

	err := db.View(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(BlocksBucket))
		currentTxTotal = bBucket.Get([]byte("t"))

		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"method": "Persist()",
		}).Panic(err)
	}

	if currentTxTotal != nil {
		currentTxTotalInt, err = strconv.ParseInt(string(currentTxTotal), 10, 64)
	} else {
		currentTxTotalInt = 0
	}

	encodedBlock := b.serialize()

	// A DB transaction to guarantee the block and [transaction] is an atom operation
	err = db.Update(func(dbtx *bolt.Tx) error {
		bBucket := dbtx.Bucket([]byte(BlocksBucket))
		txBucket := dbtx.Bucket([]byte(TransactionsBucket))

		err := bBucket.Put(b.Hash, encodedBlock)

		if err != nil {
			log.Panic(err)
		}

		for _, tx := range b.Transactions {
			// key format: blockHash_transactionId
			err := txBucket.Put(append(append(b.Hash, []byte("_")...), tx.ID...), tx.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		if isTip { // only update tip and height if this is a tip block (local or peer)
			err = bBucket.Put([]byte("l"), b.Hash)
			err = bBucket.Put([]byte("b"), []byte(fmt.Sprint(b.Height)))
		}

		err = bBucket.Put([]byte("t"), []byte(fmt.Sprint(int64(b.TotalTransactions)+currentTxTotalInt)))

		if err != nil {
			log.WithFields(log.Fields{
				"method": "Persist()",
			}).Error(err)
		}

		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"method": "Persist()",
		}).Error(err)
	}

	return b.Hash, nil
}

// DeserializeBlock deserializes a block from persistence
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.WithFields(log.Fields{
			"method": "DeserializeBlock()",
		}).Error(err)
	}

	return &block
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height uint64) *Block {
	block := &Block{time.Now().Unix(), prevBlockHash, height, []byte{}, len(transactions), transactions}
	block.Hash = block.SetHash()
	for _, tx := range transactions {
		tx.BlockHash = block.Hash
	}

	return block
}

// NewGenesisBlock creates and returns genesis block
func NewGenesisBlock(coinbase *Transaction, db *bolt.DB) *Block {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(TransactionsBucket))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"method": "NewGenesisBlock()",
		}).Error(err)
	}
	// height starts from 0
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}
