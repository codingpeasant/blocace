package blocace

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
)

// Account represents an end user's information including the public key
type Account struct {
	DateOfBirth   string `json:"dateOfBirth" validate:"len=10"`
	FirstName     string `json:"firstName" validate:"nonzero"`
	LastName      string `json:"lastName" validate:"nonzero"`
	Company       string `json:"company" validate:"nonzero"`
	Position      string `json:"position" validate:"nonzero"`
	Email         string `json:"email" validate:"min=6,max=80"`
	Phone         string `json:"phone" validate:"min=6,max=40"`
	Address       string `json:"address" validate:"min=10,max=140"`
	PublicKey     string `json:"publicKey" validate:"len=128"`
	ChallengeWord string `json:"challengeWord"`
	Role          `json:"role"`
}

// Role represents the rights of access to collections and API endpoints
type Role struct {
	Name                    string   `json:"name"`
	CollectionsWrite        []string `json:"collectionsWrite"`
	CollectionsReadOverride []string `json:"collectionsReadOverride"`
}

// Serialize serializes the transaction
func (a Account) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// DeserializeAccount deserializes an account
func DeserializeAccount(a []byte) *Account {
	var account Account

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&account)
	if err != nil {
		log.Error(err)
	}

	return &account
}

// ToMap converts an Account struct to map
func (a Account) ToMap(isAdmin bool) map[string]interface{} {
	accountMap := make(map[string]interface{})

	accountMap["dateOfBirth"] = a.DateOfBirth
	accountMap["firstName"] = a.FirstName
	accountMap["lastName"] = a.LastName
	accountMap["company"] = a.Company
	accountMap["position"] = a.Position
	accountMap["email"] = a.Email
	accountMap["phone"] = a.Phone
	accountMap["address"] = a.Address
	accountMap["publicKey"] = a.PublicKey

	if isAdmin {
		accountMap["roleName"] = a.Role.Name
		accountMap["collectionsReadOverride"] = a.Role.CollectionsReadOverride
		accountMap["collectionsWrite"] = a.Role.CollectionsWrite
	}

	return accountMap
}
