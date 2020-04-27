package blockchain

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
)

// Account represents an end user's information including the public key. Note that these account fields are just a placeholder for convenience to track identity, which doesn't affect the usage of Blocace.
// TODO: support ldap or other auth protocols
type Account struct {
	DateOfBirth   string `json:"dateOfBirth" validate:"len=10"`
	FirstName     string `json:"firstName" validate:"nonzero"`
	LastName      string `json:"lastName" validate:"nonzero"`
	Organization  string `json:"organization" validate:"nonzero"`
	Position      string `json:"position" validate:"nonzero"`
	Email         string `json:"email" validate:"min=6,max=80"`
	Phone         string `json:"phone" validate:"min=6,max=40"`
	Address       string `json:"address" validate:"min=10,max=140"`
	PublicKey     string `json:"publicKey" validate:"len=128"`
	ChallengeWord string `json:"challengeWord"`
	Role          `json:"role"`
	LastModified  int64 `json:"lastModified"`
}

// Role represents the rights of access to collections and API endpoints
type Role struct {
	Name                    string   `json:"name"`
	CollectionsWrite        []string `json:"collectionsWrite"`
	CollectionsReadOverride []string `json:"collectionsReadOverride"`
}

// Serialize serializes the account
func (a Account) Marshal() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// UnmarshalAccount deserializes an account for p2p
func UnmarshalAccount(a []byte) (Account, error) {
	var account Account

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&account)
	if err != nil {
		log.Error(err)
	}

	return account, err
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
	accountMap["organization"] = a.Organization
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
