package blockchain

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"regexp"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

const (
	BlocksBucket           = "blocks"
	TransactionsBucket     = "transactions"
	AccountsBucket         = "accounts"
	CollectionsBucket      = "collections"
	P2PPrivateKeyKey       = "p2pPrivKey"
	genesisCoinbaseRawData = `{"isActive":true,"balance":"$1,608.00","picture":"http://placehold.it/32x32","age":37,"eyeColor":"brown","name":"Rosa Sherman","gender":"male","organization":"STELAECOR","email":"rosasherman@stelaecor.com","phone":"+1 (907) 581-2115","address":"546 Meserole Street, Clara, New Jersey, 5471","about":"Reprehenderit eu pariatur proident id voluptate eu pariatur minim ut magna aliquip esse. Eu et quis sint quis et anim duis non tempor esse minim voluptate fugiat. Cillum qui nulla aute ullamco.\r\n","registered":"2018-01-15T05:53:18 +05:00","latitude":-55.183323,"longitude":-63.077504,"tags":["laborum","ex","officia","nisi","adipisicing","commodo","incididunt"],"friends":[{"id":0,"name":"Franks Harper"},{"id":1,"name":"Bettye Nash"},{"id":2,"name":"Mai Buck"}],"greeting":"Hello, Rosa Sherman! You have 3 unread messages.","favoriteFruit":"strawberry"}`

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
func IsValidSig(rawData []byte, pubKey []byte, signature []byte) bool {
	hash := crypto.Keccak256(rawData)
	if !crypto.VerifySignature(pubKey, hash[:], signature[:64]) {
		return false
	}
	return true
}

func IsValidAddress(address string) bool {
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

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	return bytes.Compare(b[i], b[j]) < 0
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// SortByteArrays sorts the slice of byte arrays
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}
