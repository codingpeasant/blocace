package p2p

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"

	"github.com/codingpeasant/blocace/blockchain"
)

// BlockP2P represents a block from a peer
type BlockP2P struct {
	PeerId            []byte
	Timestamp         int64
	PrevBlockHash     []byte
	Height            uint64
	Hash              []byte
	TotalTransactions int
	Transactions      []blockchain.Transaction
}

// Marshal serializes BlockP2P
func (b BlockP2P) Marshal() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// MapToBlock convert a BlockP2P object to blockchain.Block
func (b BlockP2P) MapToBlock() *blockchain.Block {
	var transactions []*blockchain.Transaction

	// cannot use range here because the memory address for elements being iterated is always the same one
	for i := 0; i < len(b.Transactions); i++ {
		transactions = append(transactions, &b.Transactions[i])
	}

	return &blockchain.Block{Timestamp: b.Timestamp, PrevBlockHash: b.PrevBlockHash,
		Height: b.Height, Hash: b.Hash, TotalTransactions: b.TotalTransactions, Transactions: transactions}
}

// unmarshalBlockP2P deserializes encoded bytes to BlockP2P object
func unmarshalBlockP2P(b []byte) (BlockP2P, error) {
	var blockP2P BlockP2P

	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&blockP2P)
	if err != nil {
		log.Error(err)
	}

	return blockP2P, err
}
