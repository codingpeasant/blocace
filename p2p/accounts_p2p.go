package p2p

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"

	"github.com/codingpeasant/blocace/blockchain"
)

// AccountsP2P represents all the accounts from a peer
type AccountsP2P struct {
	Accounts map[string]blockchain.Account
}

// Marshal serializes AccountsP2P
func (a AccountsP2P) Marshal() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// unmarshalAccountsP2P deserializes encoded bytes to AccountsP2P object
func unmarshalAccountsP2P(a []byte) (AccountsP2P, error) {
	var accountsP2p AccountsP2P

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&accountsP2p)
	if err != nil {
		log.Error(err)
	}

	return accountsP2p, err
}
