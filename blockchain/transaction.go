package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Transaction one transaction has on document
type Transaction struct {
	ID                 []byte // hash
	BlockHash          []byte
	PeerId             []byte // blockchain ID
	RawData            []byte
	AcceptedTimestamp  int64
	Collection         string
	PubKey             []byte
	Signature          []byte
	PermittedAddresses []string
}

// SetID sets ID of a transaction based on the raw data and timestamp
func (tx *Transaction) SetID() {
	idHash := sha256.Sum256(uuid.NewV4().Bytes())
	tx.ID = idHash[:]
}

// Serialize serializes the transaction
func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(d []byte) *Transaction {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&tx)
	if err != nil {
		log.Error(err)
	}

	return &tx
}

// Sign signs the digest of rawData
func Sign(privKey ecdsa.PrivateKey, rawDataDigest []byte) []byte {
	signature, err := crypto.Sign(rawDataDigest, &privKey)

	if err != nil {
		log.Error(err)
	}

	return signature
}

// NewTransaction creates a new transaction
func NewTransaction(peerId []byte, data []byte, collection string, pubKey []byte, signature []byte, permittedAddresses []string) *Transaction {
	tx := &Transaction{[]byte{}, []byte{}, peerId, data, time.Now().UnixNano() / 1000000, collection, pubKey, signature, permittedAddresses}
	tx.SetID()

	return tx
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(peerId []byte) *Transaction {
	return NewTransaction(peerId, []byte(genesisCoinbaseRawData), "default", []byte{}, []byte{}, nil)
}
