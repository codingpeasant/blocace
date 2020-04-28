package pool

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/blevesearch/bleve/geo"
	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"

	"github.com/codingpeasant/blocace/blockchain"
)

// Receiver represents the front door for the incoming transactions
type Receiver struct {
	transactionsBuffer     *Queue
	blockchain             *blockchain.Blockchain
	maxTxsPerBlock         int
	maxTimeToGenerateBlock int
}

// Put a transaction in JSON format to a collection. Returns isValidSig, fieldErrorMapping, transationId, error
func (r *Receiver) Put(rawData []byte, collection string, pubKey []byte, signature []byte, permittedAddresses []string) (bool, map[string]string, []byte, error) {
	isValidSig := blockchain.IsValidSig(rawData, pubKey, signature)

	if !isValidSig {
		return false, nil, nil, nil
	}

	fieldErrorMapping, err := r.checkMapping(rawData, collection)
	if err != nil {
		return true, nil, nil, err
	} else if fieldErrorMapping != nil {
		return true, fieldErrorMapping, nil, err
	}

	newTx := blockchain.NewTransaction(r.blockchain.PeerId, rawData, collection, pubKey, signature, permittedAddresses)
	r.transactionsBuffer.Append(newTx)

	return true, nil, newTx.ID, nil
}

// PutWithoutSignature a transaction in JSON format to a collection. Returns fieldErrorMapping, transationId, error
// WARNING: this makes the document unverifiable
func (r *Receiver) PutWithoutSignature(rawData []byte, collection string, permittedAddresses []string) (map[string]string, error) {
	fieldErrorMapping, err := r.checkMapping(rawData, collection)
	if err != nil {
		return nil, err
	} else if fieldErrorMapping != nil {
		return fieldErrorMapping, err
	}

	newTx := blockchain.NewTransaction(r.blockchain.PeerId, rawData, collection, nil, nil, permittedAddresses)
	r.transactionsBuffer.Append(newTx)

	return nil, nil
}

func (r *Receiver) generateBlock() {
	var candidateTxs []*blockchain.Transaction

	for i := 0; i < r.maxTxsPerBlock && r.transactionsBuffer.Length() > 0; i++ {
		tx, ok := interface{}(r.transactionsBuffer.Pop()).(*blockchain.Transaction)
		if ok {
			candidateTxs = append(candidateTxs, tx)
		}
	}

	if len(candidateTxs) > 0 {
		log.Debugf("number of txs: %d", len(candidateTxs))
		log.Debugf("queue len: %d", r.transactionsBuffer.Length())
		r.blockchain.AddBlock(candidateTxs)
	}
}

// Monitor creates a thread to monitor the transaction queue and generate block
func (r *Receiver) Monitor() {
	ticker := time.NewTicker(time.Duration(r.maxTimeToGenerateBlock) * time.Millisecond)
	log.Infof("begin to monitor transactions every %d milliseconds...\n", r.maxTimeToGenerateBlock)

	for t := range ticker.C {
		if r.transactionsBuffer.Length() > 0 {
			log.Infof("generating a block at %s...\n", t)
			r.generateBlock()
		}
	}
}

func (r Receiver) checkMapping(rawData []byte, collection string) (map[string]string, error) {
	var rawDataJSON map[string]interface{}
	err := json.Unmarshal(rawData, &rawDataJSON)

	if err != nil {
		return nil, err
	}

	var documentMapping *blockchain.DocumentMapping
	err = r.blockchain.Db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(blockchain.CollectionsBucket))

		if b == nil {
			log.WithFields(log.Fields{
				"method": "checkMapping()",
			}).Warn("bucket doesn't exist")
			return errors.New("collection doesn't exist")
		}

		encodedCollectionMapping := b.Get([]byte(collection))
		if encodedCollectionMapping == nil {
			log.WithFields(log.Fields{
				"method": "checkMapping()",
			}).Warn("collection doesn't exist")
			return errors.New("collection doesn't exist")
		}
		documentMapping = blockchain.DeserializeDocumentMapping(encodedCollectionMapping)

		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"method": "checkMapping()",
		}).Errorf("%s", err.Error())
		return nil, err
	}

	valueMaps := make(map[string]interface{})
	for field, value := range documentMapping.Fields {
		valueMap, _ := value.(map[string]interface{})
		valueMaps[field] = valueMap["type"]
	}

	validationErrors := make(map[string]string)
	for field, value := range rawDataJSON {
		switch value := value.(type) {
		case string:
			if valueMaps[field] == nil {
			} else if valueMaps[field] == "text" {
			} else if valueMaps[field] == "datetime" {
				_, err := time.Parse(time.RFC3339, value)

				if err != nil {
					validationErrors[field] = "cannot parse as RFC3339 time format"
				}
			} else {
				validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
			}
		case float64:
			if valueMaps[field] == nil {
			} else if valueMaps[field] != "number" {
				validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
			}
		case bool:
			if valueMaps[field] == nil {
			} else if valueMaps[field] != "boolean" {
				validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
			}
		case []interface{}:
			if len(value) > 0 {
				switch element := value[0].(type) {
				case string:
					if valueMaps[field] == nil {
					} else if valueMaps[field] == "text" {
					} else if valueMaps[field] == "datetime" {
						_, err := time.Parse(time.RFC3339Nano, element)

						if err != nil {
							validationErrors[field] = "cannot parse as RFC3339 time format"
						}
					} else {
						validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
					}
				case float64:
					if valueMaps[field] == nil {
					} else if valueMaps[field] != "number" {
						validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
					}
				case bool:
					if valueMaps[field] == nil {
					} else if valueMaps[field] != "boolean" {
						validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
					}
				default:
					if valueMaps[field] == nil {
					} else if valueMaps[field] == "geopoint" {
						_, _, isGeoPoint := geo.ExtractGeoPoint(value)
						if !isGeoPoint {
							validationErrors[field] = "field type should be geopoint"
						}
					} else {
						validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
					}
				}
			}
		default:
			if valueMaps[field] == nil {
			} else if valueMaps[field] == "geopoint" {
				_, _, isGeoPoint := geo.ExtractGeoPoint(value)
				if !isGeoPoint {
					validationErrors[field] = "field type should be geopoint"
				}
			} else {
				validationErrors[field] = fmt.Sprintf("field type should be %s", valueMaps[field])
			}
		}
	}

	if len(validationErrors) == 0 {
		return nil, nil
	}
	return validationErrors, nil
}

// NewReceiver creates an instance of Receiver
func NewReceiver(bc *blockchain.Blockchain, maxTxsPerBlock int, maxTimeToGenerateBlock int) *Receiver {
	return &Receiver{transactionsBuffer: NewQueue(), blockchain: bc, maxTxsPerBlock: maxTxsPerBlock, maxTimeToGenerateBlock: maxTimeToGenerateBlock}
}
