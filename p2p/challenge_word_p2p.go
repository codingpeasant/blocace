package p2p

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
)

// ChallengeWordP2P represents a challengeWord cache item from a peer
type ChallengeWordP2P struct {
	ChallengeWord string
	Address       string
}

// Marshal serializes ChallengeWordP2P
func (a ChallengeWordP2P) Marshal() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// unmarshalChallengeWordP2P deserializes encoded bytes to ChallengeWordP2P object
func unmarshalChallengeWordP2P(a []byte) (ChallengeWordP2P, error) {
	var challengeWordP2P ChallengeWordP2P

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&challengeWordP2P)
	if err != nil {
		log.Error(err)
	}

	return challengeWordP2P, err
}
