package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Error(err)
	}

	return buff.Bytes()
}

// isValidSig verifies if the rawData is a signed correctly
func isValidSig(rawData []byte, pubKey []byte, signature []byte) bool {
	hash := crypto.Keccak256(rawData)
	if !crypto.VerifySignature(pubKey, hash[:], signature[:64]) {
		return false
	}
	return true
}

func isValidAddress(address string) bool {
	r, _ := regexp.Compile("0x[a-fA-F0-9]{40}")
	return r.MatchString(address)
}

//RandStringBytesMask generates random string using masking with source
func RandStringBytesMask(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	l := len(letterBytes)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < l {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// PublicKeyToAddress generates the address of a given publicKey in bytes
func PublicKeyToAddress(publicKeyBytes []byte) (string, error) {
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)

	if err != nil {
		log.WithFields(log.Fields{
			"method": "PublicKeyToAddress()",
		}).Error(err)
		return "", err
	}
	return crypto.PubkeyToAddress(*publicKey).String(), nil
}
