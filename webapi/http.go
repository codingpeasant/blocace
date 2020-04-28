package webapi

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/boltdb/bolt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	validator "gopkg.in/validator.v2"

	"github.com/codingpeasant/blocace/blockchain"
	"github.com/codingpeasant/blocace/p2p"
	"github.com/codingpeasant/blocace/pool"
)

// HTTPHandler encapsulates the essential objects to serve http requests
type HTTPHandler struct {
	bc      *blockchain.Blockchain
	r       *pool.Receiver
	p2p     *p2p.P2P
	secret  string
	version string
}

// BlockchainInfo has current status information about the whole blockchain
type BlockchainInfo struct {
	NewestBlockID     string `json:"newestBlockId"`
	LastHeight        int    `json:"lastHeight"`
	TotalTransactions int64  `json:"totalTransactions"`
}

// BlockInfo has information about a certain block
type BlockInfo struct {
	BlockID           string `json:"blockId"`
	LastBlockID       string `json:"lastBlockId"`
	BlockHeight       uint64 `json:"blockHeight"`
	TotalTransactions int    `json:"totalTransactions"`
}

// SearchResponse determines the data in the HTTP response that the HTTP client gets
type SearchResponse struct {
	Collection string                `json:"collection"`
	Status     *bleve.SearchStatus   `json:"status"`
	Total      uint64                `json:"total_hits"`
	Hits       []blockchain.Document `json:"hits"`
}

// TransactionPayload defines the data for HTTP clients should provide to add a document to the blockchain
type TransactionPayload struct {
	RawDocument        string   `json:"rawDocument"`
	Signature          string   `json:"signature"`
	Address            string   `json:"address"`
	PermittedAddresses []string `json:"permittedAddresses"`
}

// TransactionCreationResponse has the validation information from the server to the HTTP clients
type TransactionCreationResponse struct {
	Status            string            `json:"status"`
	FieldErrorMapping map[string]string `json:"fieldErrors"`
	IsValidSignature  bool              `json:"isValidSignature"`
	TransactionID     string            `json:"transactionID"`
}

// TransactionBulkCreationResponse has the validation and count information from the server to the HTTP clients
type TransactionBulkCreationResponse struct {
	Status            string            `json:"status"`
	Total             int               `json:"total"`
	Accepted          int               `json:"accepted"`
	Dropped           int               `json:"dropped"`
	FieldErrorMapping map[string]string `json:"fieldErrors"`
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"message\": \"Blocace, the generic blockchain for all...\", \"version\": \""+h.version+"\"}")
}

// CollectionMappingCreation handles the creation of the collection (and index mapping)
func (h *HTTPHandler) CollectionMappingCreation(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, true, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// read the request body
	mappingBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"could not prcess the request body payload: "+err.Error()+"\"}", 400)
		return
	}

	newIndex, err := h.bc.Search.CreateMappingByJson(mappingBody)
	var documentMapping blockchain.DocumentMapping

	if err := json.Unmarshal(mappingBody, &documentMapping); err == nil {
		mappingToBroadcast := make(map[string]blockchain.DocumentMapping)
		mappingToBroadcast[documentMapping.Collection] = documentMapping
		h.p2p.BroadcastObject(p2p.MappingsP2P{Mappings: mappingToBroadcast})
	}

	if err != nil {
		http.Error(w, "{\"message\": \"could not create the collection: "+err.Error()+"\"}", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"collection %s created\"}", newIndex.Name())
}

// CollectionMappingGet returns the collection mapping definition
func (h HTTPHandler) CollectionMappingGet(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// find the index to operate on
	vars := mux.Vars(r)
	indexName := vars["name"]

	if nil == h.bc.Search.BlockchainIndices[indexName] {
		http.Error(w, "{\"message\": \"the collection "+indexName+" doesn't exist\"}", 404)
		return
	}

	var indexMapping *blockchain.DocumentMapping
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.CollectionsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleCollectionMappingGet",
				"address": r.Header.Get("address"),
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedIndexMapping := b.Get([]byte(indexName))

		if encodedIndexMapping == nil {
			log.WithFields(log.Fields{
				"route":   "HandleCollectionMappingGet",
				"address": r.Header.Get("address"),
			}).Error("collection doesn't exist")
			return errors.New("collection doesn't exist")
		}
		indexMapping = blockchain.DeserializeDocumentMapping(encodedIndexMapping)

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"could not find the collection\"}", 404)
		return
	}

	rv := struct {
		Message string                     `json:"message"`
		Mapping blockchain.DocumentMapping `json:"mapping"`
	}{
		Message: "ok",
		Mapping: *indexMapping,
	}

	mustEncode(w, rv)
}

// CollectionList returns the all the collection names
func (h HTTPHandler) CollectionList(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	var indexNames []string
	for name := range h.bc.Search.BlockchainIndices {
		indexNames = append(indexNames, name)
	}

	rv := struct {
		Message string   `json:"message"`
		Indexes []string `json:"collections"`
	}{
		Message: "ok",
		Indexes: indexNames,
	}

	mustEncode(w, rv)
}

// HandleInfo returns the basic information of the blockchain
func (h HTTPHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	var lastHeight int
	var totalTransactionsInt int64

	h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.BlocksBucket))
		encodedBlock := b.Get(h.bc.Tip)
		block := blockchain.DeserializeBlock(encodedBlock)
		lastHeight = int(block.Height)

		totalTransactions := b.Get([]byte("t"))
		totalTransactionsInt, _ = strconv.ParseInt(string(totalTransactions), 10, 64)

		return nil
	})

	blockchainInfoJSON, err := json.Marshal(BlockchainInfo{NewestBlockID: fmt.Sprintf("%x", h.bc.Tip), LastHeight: lastHeight, TotalTransactions: totalTransactionsInt})

	if err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleInfo",
			"address": r.Header.Get("address"),
		}).Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(blockchainInfoJSON)
}

// HandleTransaction put and index new transaction
// {
//     "rawDocument": "{\"id\":\"10001\",\"message\":\"Send 10000 BTC to Ivan\"}",
//     "signature": "8e0063b76c2aed4982e1b62c713b0a7cf74f2b548b8c032659da65404c3d0b9777b8f8613f3e87e43680ec638949e263658ef5608bad7359e1075e285f49dd8d",
//     "permittedAddresses" : ["0x07322C5A59047c09e87C284503F64f7FdDD17aBd", "0x931D387731bBbC988B312206c74F77D004D6B84b"]
// }
func (h *HTTPHandler) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// find the index to operate on
	vars := mux.Vars(r)
	indexName := vars["collection"]

	if nil == h.bc.Search.BlockchainIndices[indexName] {
		http.Error(w, "{\"message\": \"no such collection: "+indexName+"\"}", 404)
		return
	}

	// check writing permission
	address := r.Header.Get("address")
	var account *blockchain.Account
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))
		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleTransaction",
				"address": address,
			}).Info("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))
		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "HandleTransaction",
				"address": address,
			}).Info("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		return nil
	})

	if !funk.ContainsString(account.CollectionsWrite, indexName) {
		log.WithFields(log.Fields{
			"route":   "HandleTransaction",
			"address": address,
		}).Info("insufficient permission to write to collection: ", indexName)
		http.Error(w, "{\"message\": \"insufficient permission to write to collection: "+indexName+"\"}", 401)
		return
	}

	// read the request body
	transactionBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"error reading the request body: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var transactionPayload TransactionPayload
	err = json.Unmarshal(transactionBody, &transactionPayload)
	if err != nil {
		http.Error(w, "{\"message\": \"error parsing the payload: "+err.Error()+"\"}", 400)
		return
	}

	var publicKey []byte
	publicKey, err = hex.DecodeString(account.PublicKey)
	if err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleTransaction",
			"address": r.Header.Get("address"),
		}).Error("hex.DecodeString publicKey: " + err.Error())
		return
	}

	signatureBytes, err := hex.DecodeString(transactionPayload.Signature)
	if err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleTransaction",
			"address": r.Header.Get("address"),
		}).Error("couldn't recognize the publicKey: " + err.Error())
		http.Error(w, "{\"message\": \"couldn't recognize the publicKey: "+err.Error()+"\"}", 500)
		return
	}

	transactionPayload.PermittedAddresses = append(transactionPayload.PermittedAddresses, r.Header.Get("address")) // add self
	isValidSig, fieldErrorMapping, txID, err := h.r.Put([]byte(transactionPayload.RawDocument), indexName, publicKey, signatureBytes, transactionPayload.PermittedAddresses)
	if err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleTransaction",
			"address": r.Header.Get("address"),
		}).Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		mustEncode(w, TransactionCreationResponse{Status: "internal error"})
		return
	}

	if !isValidSig {
		w.WriteHeader(http.StatusBadRequest)
		mustEncode(w, TransactionCreationResponse{Status: "bad signature", IsValidSignature: false})
		return
	}

	if fieldErrorMapping != nil {
		w.WriteHeader(http.StatusBadRequest)
		mustEncode(w, TransactionCreationResponse{Status: "field validation failed", IsValidSignature: true, FieldErrorMapping: fieldErrorMapping})
		return
	}

	mustEncode(w, TransactionCreationResponse{Status: "ok", IsValidSignature: true, TransactionID: fmt.Sprintf("%x", txID)})
}

// HandleTransactionBulk put and index new transactions in bulk. Transactions payload don't need signatures.
// [
//     {
//         "id": "10001",
//         "message": "Send 1 BTC to Ivan"
//     },
//     {
//         "id": "10002",
//         "message": "Send 2 BTC to Ivan"
//     },
//     {
//         "id": "10003",
//         "message": "Send 3 BTC to Ivan"
//     },
//     {
//         "id": "10004",
//         "message": "Send 4 BTC to Ivan"
//     }
// ]
// WARNING: this makes the document unverifiable
func (h HTTPHandler) HandleTransactionBulk(w http.ResponseWriter, r *http.Request) {
	// find the index to operate on
	vars := mux.Vars(r)
	indexName := vars["collection"]

	if nil == h.bc.Search.BlockchainIndices[indexName] {
		http.Error(w, "{\"message\": \"no such collection: "+indexName+"\"}", 404)
		return
	}

	// read the request body
	transactionBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"error reading the request body: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var jsonDocs []map[string]interface{}
	err = json.Unmarshal(transactionBody, &jsonDocs)
	if err != nil {
		http.Error(w, "{\"message\": \"error parsing the payload: "+err.Error()+"\"}", 400)
		return
	}

	accepted := 0
	for _, transaction := range jsonDocs {
		jsonBytes, err := json.Marshal(transaction)

		if err != nil {
			log.WithFields(log.Fields{
				"route":   "HandleTransaction",
				"address": r.Header.Get("address"),
			}).Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			mustEncode(w, TransactionBulkCreationResponse{Status: "cannot parse json payload", Total: len(jsonDocs), Accepted: accepted, Dropped: (len(jsonDocs) - accepted)})
			return
		}

		fieldErrorMapping, err := h.r.PutWithoutSignature(jsonBytes, indexName, nil)

		if err != nil {
			log.WithFields(log.Fields{
				"route":   "HandleTransaction",
				"address": r.Header.Get("address"),
			}).Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			mustEncode(w, TransactionBulkCreationResponse{Status: err.Error(), Total: len(jsonDocs), Accepted: accepted, Dropped: (len(jsonDocs) - accepted)})
			return
		}

		if fieldErrorMapping != nil {
			w.WriteHeader(http.StatusBadRequest)
			mustEncode(w, TransactionBulkCreationResponse{Status: "field validation failed", Total: len(jsonDocs), Accepted: accepted, Dropped: (len(jsonDocs) - accepted), FieldErrorMapping: fieldErrorMapping})
			return
		}
		accepted++
	}

	w.WriteHeader(http.StatusAccepted)
	mustEncode(w, TransactionBulkCreationResponse{Status: "ok", Total: len(jsonDocs), Accepted: accepted, Dropped: (len(jsonDocs) - accepted)})
}

// AccountRegistration register the account information
// {
//     "dateOfBirth": "2018-10-01",
//     "firstName": "Hooper",
//     "lastName": "Vincent",
//     "organization": "MITROC",
//     "position": "VP of Marketing",
//     "email": "hoopervincent@mitroc.com",
//     "phone": "+1 (849) 503-2756",
//     "address": "699 Canton Court, Mulino, South Dakota, 9647",
//     "publicKey":"e4a15344314a15c70a47e18fadc8117939a6dc5ed863ced84a898694b241d10fa129eff3989ec98393c52bac6d86d0d72534061538eb1e513aaae4def5f83fbb"
// }
func (h *HTTPHandler) AccountRegistration(w http.ResponseWriter, r *http.Request) {
	// read the request body
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"could not process the request body payload: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var account blockchain.Account
	if err = json.Unmarshal(requestBody, &account); err != nil {
		http.Error(w, "{\"message\": \"error parsing the json payload: "+err.Error()+"\"}", 400)
		return
	}

	if err = validator.Validate(account); err != nil {
		http.Error(w, "{\"message\": \"error validating the field: "+err.Error()+"\"}", 400)
		return
	}

	account.PublicKey = "04" + account.PublicKey // appending 04 to be compatible with ecdsa.PublicKey uncompressed form
	account.Role.Name = "user"                   // user only registration
	account.Role.CollectionsWrite = nil          // don't allow setting permissions
	account.Role.CollectionsReadOverride = nil
	account.LastModified = time.Now().UnixNano() / 1000000
	publicKeyBytes, err := hex.DecodeString(account.PublicKey)
	if err != nil {
		http.Error(w, "{\"message\": \"error parsing public key: "+err.Error()+"\"}", 400)
		return
	}

	var address string
	if address, err = blockchain.PublicKeyToAddress(publicKeyBytes); err != nil {
		http.Error(w, "{\"message\": \"error parsing public key: "+err.Error()+"\"}", 400)
		return
	}

	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleAccountRegistration",
				"address": address,
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))

		if encodedAccount != nil {
			log.WithFields(log.Fields{
				"route":   "HandleAccountRegistration",
				"address": address,
			}).Error("account exists already")
			return errors.New("account exists already")
		}

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"account exists already\"}", 400)
		return
	}

	if err = h.bc.RegisterAccount([]byte(address), account); err != nil {
		http.Error(w, "{\"message\": \"error adding the account: "+err.Error()+"\"}", 400)
		return
	}

	// broadcast to peers
	accountMap := make(map[string]blockchain.Account)
	accountMap[address] = account
	h.p2p.BroadcastObject(p2p.AccountsP2P{Accounts: accountMap})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"account created\", \"address\": \"%s\"}", address)
}

// AccountUpdate updates an existing account's information
// {
// 	"dateOfBirth": "2018-10-01",
// 	"firstName": "Hooper",
// 	"lastName": "Vincent",
// 	"organization": "MITROC",
//  "position": "VP of Marketing",
// 	"email": "hoopervincent@mitroc.com",
// 	"phone": "+1 (849) 503-2756",
// 	"address": "699 Canton Court, Mulino, South Dakota, 9647"
// }
func (h *HTTPHandler) AccountUpdate(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if address != r.Header.Get("address") {
		http.Error(w, "{\"message\": \"you can only update your own account\"}", 401)
		return
	}

	// read the request body
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"could not process the request body payload: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var account blockchain.Account
	if err = json.Unmarshal(requestBody, &account); err != nil {
		http.Error(w, "{\"message\": \"error parsing the json payload: "+err.Error()+"\"}", 400)
		return
	}

	if err = validator.Validate(account); err != nil {
		http.Error(w, "{\"message\": \"error validating the field: "+err.Error()+"\"}", 400)
		return
	}

	var oldAccount *blockchain.Account
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))
		if b == nil {
			log.WithFields(log.Fields{
				"route":   "AccountUpdate",
				"address": address,
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))
		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "AccountUpdate",
				"address": address,
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		oldAccount = blockchain.DeserializeAccount(encodedAccount)

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"account doesn't exist\"}", 404)
		return
	}

	account.PublicKey = oldAccount.PublicKey
	account.Role = oldAccount.Role
	account.LastModified = time.Now().UnixNano() / 1000000

	if err = h.bc.RegisterAccount([]byte(address), account); err != nil {
		http.Error(w, "{\"message\": \"error adding the account: "+err.Error()+"\"}", 400)
		return
	}

	// broadcast to peers
	accountMap := make(map[string]blockchain.Account)
	accountMap[address] = account
	h.p2p.BroadcastObject(p2p.AccountsP2P{Accounts: accountMap})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"account updated\", \"address\": \"%s\"}", address)
}

// AccountGet returns an account's information for a given address
func (h HTTPHandler) AccountGet(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if !blockchain.IsValidAddress(address) {
		http.Error(w, "{\"message\": \"not a valid address\"}", 400)
		return
	}

	var account *blockchain.Account

	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "AccountGet",
				"address": address,
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))

		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "AccountGet",
				"address": address,
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"account doesn't exist\"}", 404)
		return
	}

	mustEncode(w, account.ToMap(r.Header.Get("role") == "admin"))
}

// SetAccountReadWrite set the user's collection-level read override and write permission
// {
// 	"collectionsWrite": ["default", "collection1"],
// 	"collectionsReadOverride": ["default", "collection2"]
// }
func (h HTTPHandler) SetAccountReadWrite(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, true, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if !blockchain.IsValidAddress(address) {
		http.Error(w, "{\"message\": \"not a valid address\"}", 400)
		return
	}

	// read the request body
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"could not process the request body payload: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var role blockchain.Role
	if err = json.Unmarshal(requestBody, &role); err != nil {
		http.Error(w, "{\"message\": \"error parsing the json payload: "+err.Error()+"\"}", 400)
		return
	}

	var account *blockchain.Account
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "SetAccountReadWrite",
				"address": address,
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))
		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "SetAccountReadWrite",
				"address": address,
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"account doesn't exist\"}", 404)
		return
	}

	roleName := account.Role.Name
	account.Role = role
	account.Role.Name = roleName

	if err = h.bc.RegisterAccount([]byte(address), *account); err != nil {
		http.Error(w, "{\"message\": \"error adding the account: "+err.Error()+"\"}", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"account permission updated\", \"address\": \"%s\"}", address)
}

// HandleBlockInfo returns an account's information for a given address
func (h HTTPHandler) HandleBlockInfo(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// find the index to operate on
	vars := mux.Vars(r)
	blockID, err := hex.DecodeString(vars["blockId"])

	if err != nil {
		http.Error(w, "{\"message\": \"invalid block ID\"}", 400)
		return
	}

	var block *blockchain.Block

	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.BlocksBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleBlockInfo",
				"address": r.Header.Get("address"),
			}).Warn("bucket doesn't exist")
			return errors.New("block doesn't exist")
		}

		encodedBlock := b.Get(blockID)

		if encodedBlock == nil {
			log.WithFields(log.Fields{
				"route":   "HandleBlockInfo",
				"address": r.Header.Get("address"),
			}).Error("block doesn't exist")
			return errors.New("block doesn't exist")
		}
		block = blockchain.DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"block doesn't exist\"}", 404)
		return
	}

	blockInfoResponse := BlockInfo{BlockID: fmt.Sprintf("%x", block.Hash), LastBlockID: fmt.Sprintf("%x", block.PrevBlockHash), BlockHeight: block.Height, TotalTransactions: block.TotalTransactions}

	mustEncode(w, blockInfoResponse)
}

// HandleMerklePath returns the necessary transaction hashes for clients to verify if a transaction has been included in the block
func (h HTTPHandler) HandleMerklePath(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// find the index to operate on
	vars := mux.Vars(r)
	blockID, err := hex.DecodeString(vars["blockId"])
	txID, err := hex.DecodeString(vars["txId"])

	if err != nil {
		http.Error(w, "{\"message\": \"invalid block transaction ID\"}", 400)
		return
	}

	var block *blockchain.Block

	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.BlocksBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleMerklePath",
				"address": r.Header.Get("address"),
			}).Warn("block bucket doesn't exist")
			return errors.New("block doesn't exist")
		}

		encodedBlock := b.Get(blockID)

		if encodedBlock == nil {
			log.WithFields(log.Fields{
				"route":   "HandleMerklePath",
				"address": r.Header.Get("address"),
			}).Error("block doesn't exist")
			return errors.New("block doesn't exist")
		}
		block = blockchain.DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"block doesn't exist\"}", 404)
		return
	}

	h.bc.Db.View(func(dbtx *bolt.Tx) error {
		// Assume bucket exists and has keys
		c := dbtx.Bucket([]byte(blockchain.TransactionsBucket)).Cursor()

		prefix := block.Hash
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			block.Transactions = append(block.Transactions, blockchain.DeserializeTransaction(v))
		}

		return nil
	})

	verificationPath := block.GetMerkleTree().GetVerificationPath(txID)
	if verificationPath == nil {
		http.Error(w, "{\"message\": \"couldn't create the merkle tree for this transation\"}", 400)
		return
	}

	verificationPathString := make(map[int]string)
	for index, hash := range verificationPath {
		verificationPathString[index] = fmt.Sprintf("%x", hash)
	}

	rv := struct {
		Status     string         `json:"status"`
		MerklePath map[int]string `json:"verificationPath"`
	}{
		Status:     "ok",
		MerklePath: verificationPathString,
	}

	mustEncode(w, rv)
}

// HandleSearch handles the search queries against the search engine
// {
// 	"size": 10,
// 	"explain": true,
// 	"highlight": {},
// 	"query": {
// 		"boost": 1,
// 		"match": "Canada",
// 		"field": "country",
// 		"prefix_length": 0,
// 		"fuzziness": 0
// 	}
// }
func (h HTTPHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	err := processJWT(r, false, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \""+err.Error()+"\"}", 401)
		return
	}

	// find the index to operate on
	vars := mux.Vars(r)
	indexName := vars["collection"]

	if nil == h.bc.Search.BlockchainIndices[indexName] {
		http.Error(w, "{\"message\": \"no such collection: "+indexName+"\"}", 404)
		return
	}

	// read the request body
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"err reading the request body: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var searchRequest bleve.SearchRequest
	err = json.Unmarshal(requestBody, &searchRequest)
	if err != nil {
		http.Error(w, "{\"message\": \"error parsing the query: "+err.Error()+"\"}", 400)
		return
	}

	// check read overriding permission
	address := r.Header.Get("address")
	var account *blockchain.Account
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))
		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleSearch",
				"address": address,
			}).Warn("bucket doesn't exist")
			return errors.New("bucket doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))
		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "HandleSearch",
				"address": address,
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		return nil
	})

	if !funk.ContainsString(account.CollectionsReadOverride, indexName) {
		// only addresses in _permittedAddresses can access
		permittedAddressesQuery := query.NewMatchQuery(r.Header.Get("address"))
		permittedAddressesQuery.SetField("_permittedAddresses")

		conjunctQuery := query.NewConjunctionQuery([]query.Query{searchRequest.Query, permittedAddressesQuery})
		searchRequest.Query = conjunctQuery
	}

	// validate the query
	if srqv, ok := searchRequest.Query.(query.ValidatableQuery); ok {
		err = srqv.Validate()
		if err != nil {
			log.WithFields(log.Fields{
				"route":   "HandleSearch",
				"address": r.Header.Get("address"),
			}).Errorf("error validating the query: %s. The query: %s", err.Error(), searchRequest.Query)
			http.Error(w, "{\"message\": \"error validating the query: "+err.Error()+"\"}", 400)
			return
		}
	}

	searchRequest.Explain = false
	// execute the query
	searchResponse, err := h.bc.Search.BlockchainIndices[indexName].Search(&searchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleSearch",
			"address": r.Header.Get("address"),
		}).Error("error executing query: " + err.Error())
		http.Error(w, "error executing query: "+err.Error(), 500)
		return
	}

	hits := []blockchain.Document{}
	for _, hit := range searchResponse.Hits {
		h.bc.Db.View(func(dbtx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := dbtx.Bucket([]byte(blockchain.TransactionsBucket))

			v := b.Get([]byte(hit.ID))

			tx := blockchain.DeserializeTransaction(v)

			if err != nil {
				log.WithFields(log.Fields{
					"route":   "HandleSearch",
					"address": r.Header.Get("address"),
				}).Error("error encoding doc: ", err.Error())
				return err
			}

			var publicKey *ecdsa.PublicKey
			var transactionAddress string
			if tx.PubKey != nil {
				if publicKey, err = crypto.UnmarshalPubkey(tx.PubKey); err != nil {
					log.WithFields(log.Fields{
						"route":   "HandleSearch",
						"address": r.Header.Get("address"),
					}).Error("error unmarshal public key bytes: ", err.Error())
					return err
				}
				transactionAddress = crypto.PubkeyToAddress(*publicKey).String()
			} else {
				transactionAddress = ""
			}

			hits = append(hits, blockchain.Document{ID: fmt.Sprintf("%x", tx.ID), BlockID: fmt.Sprintf("%x", tx.BlockHash), BlockchainId: fmt.Sprintf("%x", tx.PeerId), Source: fmt.Sprintf("%s", tx.RawData), Timestamp: time.Unix(0, tx.AcceptedTimestamp*int64(time.Millisecond)).Format(time.RFC3339Nano), Signature: fmt.Sprintf("%x", tx.Signature), Address: transactionAddress})

			return nil
		})
	}

	mustEncode(w, SearchResponse{Collection: indexName, Status: searchResponse.Status, Total: searchResponse.Total, Hits: hits})
}

// HandleJWT checks the credentials and return corresponding JWT
// {
// 	"address": "0x07322C5A59047c09e87C284503F64f7FdDD17aBd",
// 	"signature": "6b2064ddf73f7b96559ecae424b3b657d1daf62078305e92af991c22e04808d476e8161ec7be58324b662042965a9935de8fb697eb3df7afad5d1885f129f666"
// }
func (h HTTPHandler) HandleJWT(w http.ResponseWriter, r *http.Request) {
	// read the request body
	transactionBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"message\": \"error reading the request body: "+err.Error()+"\"}", 400)
		return
	}

	// parse the request
	var authPayload TransactionPayload // reuse TransactionPayload for simplicity
	err = json.Unmarshal(transactionBody, &authPayload)
	if err != nil {
		http.Error(w, "{\"message\": \"error parsing the payload: "+err.Error()+"\"}", 400)
		return
	}

	var account *blockchain.Account
	err = h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWT",
				"address": r.Header.Get("address"),
			}).Warn("bucket doesn't exist")
			return errors.New("account doesn't exist")
		}

		encodedAccount := b.Get([]byte(authPayload.Address))

		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWT",
				"address": r.Header.Get("address"),
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		if account == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWT",
				"address": r.Header.Get("address"),
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"couldn't find the account: "+err.Error()+"\"}", 404)
		return
	}

	signatureBytes, err := hex.DecodeString(authPayload.Signature)
	if err != nil {
		http.Error(w, "{\"message\": \"couldn't process the signature: "+err.Error()+"\"}", 400)
		return
	}

	if len(account.ChallengeWord) == 0 {
		http.Error(w, "{\"message\": \"no challenge word available\"}", 404)
		return
	}

	publicKey, err := hex.DecodeString(account.PublicKey)
	if err != nil {
		http.Error(w, "{\"message\": \"public key not valid\"}", 500)
		return
	}

	isValidSig := blockchain.IsValidSig([]byte(account.ChallengeWord), publicKey, signatureBytes)
	if !isValidSig {
		http.Error(w, "{\"message\": \"signature invalid\"}", 400)
		return
	}

	token, err := issueToken(authPayload.Address, account.Role, h.secret)
	if err != nil {
		http.Error(w, "{\"message\": \"couldn't issue jwt: "+err.Error()+"\"}", 500)
		return
	}

	account.ChallengeWord = ""
	if err = h.bc.RegisterAccount([]byte(authPayload.Address), *account); err != nil {
		log.WithFields(log.Fields{
			"route":   "HandleJWT",
			"address": r.Header.Get("address"),
		}).Error("error reset challenge word: " + err.Error())
		http.Error(w, "{\"message\": \"internal error: "+err.Error()+"\"}", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"JWT issued\", \"token\": \"%s\"}", token)
}

// JWTChallenge creates a random challenge word for the clients to sign
func (h HTTPHandler) JWTChallenge(w http.ResponseWriter, r *http.Request) {
	// find the index to operate on
	vars := mux.Vars(r)
	address := vars["address"]

	if !blockchain.IsValidAddress(address) {
		http.Error(w, "{\"message\": \"not a valid address\"}", 400)
		log.WithFields(log.Fields{
			"route":   "JWTChallenge",
			"address": r.Header.Get("address"),
		}).Warn("not a valid address")
		return
	}

	var account *blockchain.Account
	err := h.bc.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.AccountsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWTChallenge",
				"address": r.Header.Get("address"),
			}).Warn("bucket doesn't exist")
			return errors.New("account doesn't exist")
		}

		encodedAccount := b.Get([]byte(address))

		if encodedAccount == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWTChallenge",
				"address": r.Header.Get("address"),
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}
		account = blockchain.DeserializeAccount(encodedAccount)

		if account == nil {
			log.WithFields(log.Fields{
				"route":   "HandleJWTChallenge",
				"address": r.Header.Get("address"),
			}).Error("account doesn't exist")
			return errors.New("account doesn't exist")
		}

		return nil
	})

	if err != nil {
		http.Error(w, "{\"message\": \"couldn't find the account: "+err.Error()+"\"}", 404)
		return
	}

	account.ChallengeWord = blockchain.RandStringBytesMask(64)
	if err = h.bc.RegisterAccount([]byte(address), *account); err != nil {
		http.Error(w, "{\"message\": \"error updating the challenge word: "+err.Error()+"\"}", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"message\": \"challenge word created\", \"challenge\": \"%s\"}", account.ChallengeWord)
}

// ErrorHandler handles non-existing route
func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"message\": \"handler not found for path: " + r.URL.Path + "\"}"))
}

func mustEncode(w io.Writer, i interface{}) {
	if headered, ok := w.(http.ResponseWriter); ok {
		headered.Header().Set("Cache-Control", "no-cache")
		headered.Header().Set("Content-type", "application/json")
	}

	e := json.NewEncoder(w)
	if err := e.Encode(i); err != nil {
		log.WithFields(log.Fields{
			"route": "mustEncode",
		}).Error("error encoding response to json: ", err.Error())
	}
}

func processJWT(r *http.Request, requireAdmin bool, secret string) error {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) == 0 {
		return errors.New("missing authorization header")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	claims, err := verifyToken(tokenString, secret)
	if err != nil {
		return err
	}

	address := claims.(jwt.MapClaims)["address"].(string)
	roleName := claims.(jwt.MapClaims)["roleName"].(string)

	if requireAdmin && roleName != "admin" {
		return errors.New("insufficient permission")
	}

	r.Header.Set("address", address)
	r.Header.Set("role", roleName)

	return nil
}

// NewHTTPHandler create a new instance of HTTPHandler
func NewHTTPHandler(b *blockchain.Blockchain, r *pool.Receiver, p *p2p.P2P, secret string, version string) HTTPHandler {
	return HTTPHandler{b, r, p, secret, version}
}
