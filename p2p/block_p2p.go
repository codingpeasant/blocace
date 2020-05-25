package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"

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
	IsTip             bool
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

// MapToBlock verified each transaction in the BlockP2P on wire and convert it to blockchain.Block
func (b BlockP2P) MapToBlock() (*blockchain.Block, error) {
	var transactions []*blockchain.Transaction

	// cannot use range here because the memory address for elements being iterated is always the same one
	for i := 0; i < len(b.Transactions); i++ {
		if bytes.Compare(b.Transactions[i].Signature, []byte{}) == 0 { // skipping for genisis transation and bulk loading
			transactions = append(transactions, &b.Transactions[i])
		} else if blockchain.IsValidSig(b.Transactions[i].RawData, b.Transactions[i].PubKey, b.Transactions[i].Signature) {
			transactions = append(transactions, &b.Transactions[i])
		} else {
			transactions = nil
			return nil, fmt.Errorf("transaction signature verification failed, abandon this block")
		}
	}

	return &blockchain.Block{Timestamp: b.Timestamp, PrevBlockHash: b.PrevBlockHash,
		Height: b.Height, Hash: b.Hash, TotalTransactions: b.TotalTransactions, Transactions: transactions}, nil
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
