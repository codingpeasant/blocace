package p2p

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"

	"github.com/codingpeasant/blocace/blockchain"
)

// MappingsP2P represents all the collection mappings (schemas) from a peer
type MappingsP2P struct {
	Mappings map[string]blockchain.DocumentMapping
}

// Marshal serializes MappingsP2P
func (a MappingsP2P) Marshal() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// unmarshalMappingsP2P deserializes encoded bytes to MappingsP2P object
func unmarshalMappingsP2P(a []byte) (MappingsP2P, error) {
	var mappingsP2p MappingsP2P

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&mappingsP2p)
	if err != nil {
		log.Error(err)
	}

	return mappingsP2p, err
}
