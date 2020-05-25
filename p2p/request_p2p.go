package p2p

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
)

const accountsRequestType = "accounts" // address:lastModified
const mappingsRequestType = "mappings" // collecionName:collectionName
const blockRequestType = "block"       // peerId:blockId or local:[tip or blockId] (don't support multiple key-value pairs yet)

// RequestP2P represents common p2p request body
type RequestP2P struct {
	RequestType       string
	RequestParameters map[string]string
}

// Marshal serializes RequestP2P
func (a RequestP2P) Marshal() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// unmarshalRequestP2P deserializes encoded bytes to RequestP2P object
func unmarshalRequestP2P(a []byte) (RequestP2P, error) {
	var requestP2p RequestP2P

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&requestP2p)
	if err != nil {
		log.Error(err)
	}

	return requestP2p, err
}
