package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/blevesearch/bleve/geo"
	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
)

// Receiver represents the front door for the incoming transactions
type Receiver struct {
	transactionsBuffer *Queue
	blockchain         *Blockchain
}

// Put a transaction in JSON format to a collection. Returns isValidSig, fieldErrorMapping, transationId, error
func (r *Receiver) Put(rawData []byte, collection string, pubKey []byte, signature []byte, permittedAddresses []string) (bool, map[string]string, []byte, error) {
	isValidSig := isValidSig(rawData, pubKey, signature)

	if !isValidSig {
		return false, nil, nil, nil
	}

	fieldErrorMapping, err := r.checkMapping(rawData, collection)
	if err != nil {
		return true, nil, nil, err
	} else if fieldErrorMapping != nil {
		return true, fieldErrorMapping, nil, err
	}

	newTx := NewTransaction(rawData, collection, pubKey, signature, permittedAddresses)
	r.transactionsBuffer.Add(newTx)

	return true, nil, newTx.ID, nil
}

func (r *Receiver) generateBlock() {
	var candidateTxs []*Transaction

	for i := 0; i < maxTxsPerBlock && r.transactionsBuffer.Length() > 0; i++ {
		candidateTxs = append(candidateTxs, r.transactionsBuffer.Remove())
	}

	candidateTxs = removeDuplicateTransactions(candidateTxs)
	if len(candidateTxs) > 0 {
		r.blockchain.AddBlock(candidateTxs)
	}
}

// Monitor creates a thread to monitor the transaction queue and generate block
func (r *Receiver) Monitor() {
	ticker := time.NewTicker(time.Duration(maxTimeToGenerateBlock) * time.Millisecond)
	log.Infof("begin to monitor transactions every %d milliseconds...\n", maxTimeToGenerateBlock)

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

	var documentMapping *DocumentMapping
	err = r.blockchain.db.View(func(dbtx *bolt.Tx) error {
		b := dbtx.Bucket([]byte(collectionsBucket))

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
		documentMapping = DeserializeDocumentMapping(encodedCollectionMapping)

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
func NewReceiver(bc *Blockchain) *Receiver {
	return &Receiver{transactionsBuffer: NewQueue(), blockchain: bc}
}
