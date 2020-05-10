package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
)

const indexDefault = "default"

// Search encapsulates all the indices with search engine features
type Search struct {
	sync.Mutex
	db                *bolt.DB
	indexDirRoot      string
	BlockchainIndices map[string]bleve.Index
}

// Document represents a document with metadata in the search result
type Document struct {
	ID           string `json:"_id"`
	BlockID      string `json:"_blockId"`
	BlockchainId string `json:"_blockchainId"` // peerId
	Source       string `json:"_source"`
	Timestamp    string `json:"_timestamp"`
	Signature    string `json:"_signature"`
	Address      string `json:"_address"` // Issuer address
}

// NewSearch create an instance to access the search features
func NewSearch(db *bolt.DB, dataDir string) (*Search, error) {
	blockchainIndices := make(map[string]bleve.Index)

	indexDirRoot := dataDir + filepath.Dir("/") + "collections"
	defaultIndex, err := bleve.Open(indexDirRoot + "/" + indexDefault)

	if err != nil {
		log.Infof("%s: %s. creating the default collection instead...\n", err, indexDirRoot)

		jsonSchema := `
		{
			"collection": "default",
			"fields": {
				"id": {"type": "number"},
				"message": {"type": "text"}
			}
		}
		`

		search := Search{db: db, indexDirRoot: indexDirRoot, BlockchainIndices: blockchainIndices}

		defaultIndex, err := search.CreateMappingByJson([]byte(jsonSchema))

		if err != nil {
			log.Fatalf("creating default collection err: %s", err)
			return nil, err
		}
		search.BlockchainIndices[indexDefault] = defaultIndex
		return &search, nil
	}

	log.Info("opening existing collections...")
	blockchainIndices["default"] = defaultIndex

	files, err := ioutil.ReadDir(indexDirRoot)
	if err != nil {
		log.Fatal(err)
	}

	// add all other indices than the default
	for _, file := range files {
		if file.Name() != indexDefault {
			bleveIndex, err := bleve.Open(indexDirRoot + filepath.Dir("/") + file.Name())

			if err != nil {
				log.Error("cannot add this index: " + err.Error())
			} else {
				blockchainIndices[file.Name()] = bleveIndex
			}
		}
	}

	return &Search{db: db, indexDirRoot: indexDirRoot, BlockchainIndices: blockchainIndices}, nil
}

// IndexBlock index all the txs in a block
func (s *Search) IndexBlock(block *Block, peerId []byte) {
	s.Lock()
	defer s.Unlock()
	var jsonDoc map[string]interface{}

	// using batch index for better performance
	indexBatches := make(map[string]*bleve.Batch)
	for collection, index := range s.BlockchainIndices {
		indexBatches[collection] = index.NewBatch()
	}

	for _, tx := range block.Transactions {
		// do not index the doc where there is no index exists for it
		if nil == s.BlockchainIndices[tx.Collection] {
			//log.Println("The collection " + tx.Collection + " doesn't exist... Skipped the indexing.")
			continue
		}

		// parse bytes as json
		err := json.Unmarshal(tx.RawData, &jsonDoc)

		if err != nil {
			log.Errorf("error indexing tx with ID %x: %s", tx.ID, err)
		}

		// all searchable system fields
		jsonDoc["_type"] = tx.Collection
		jsonDoc["_blockId"] = fmt.Sprintf("%x", tx.BlockHash)
		jsonDoc["_timestamp"] = time.Unix(0, tx.AcceptedTimestamp*int64(time.Millisecond)).Format(time.RFC3339)
		jsonDoc["_publicKey"] = fmt.Sprintf("%x", tx.PubKey)
		jsonDoc["_id"] = fmt.Sprintf("%x", tx.ID)
		jsonDoc["_peerId"] = fmt.Sprintf("%x", peerId)
		jsonDoc["_permittedAddresses"] = tx.PermittedAddresses

		indexBatches[tx.Collection].Index(string(append(append(block.Hash, []byte("_")...), tx.ID...)), jsonDoc)
	}

	for collection, batch := range indexBatches {
		s.BlockchainIndices[collection].Batch(batch)
	}
}

// DocumentMapping represents the schema of a collection
type DocumentMapping struct {
	Collection string                 `json:"collection"`
	Fields     map[string]interface{} `json:"fields"`
}

// Serialize serializes the transaction
func (dm DocumentMapping) Serialize() []byte {
	var result bytes.Buffer

	mappingExpression := map[string]interface{}{
		"id": "{\"type\": \"text\"}",
	}
	gob.Register(mappingExpression)

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(dm)
	if err != nil {
		log.Error(err)
	}

	return result.Bytes()
}

// DeserializeDocumentMapping deserializes encoded bytes to an DocumentMapping object
func DeserializeDocumentMapping(a []byte) *DocumentMapping {
	var dm DocumentMapping

	mappingExpression := map[string]interface{}{
		"id": "{\"type\": \"text\"}",
	}
	gob.Register(mappingExpression)

	decoder := gob.NewDecoder(bytes.NewReader(a))
	err := decoder.Decode(&dm)
	if err != nil {
		log.Error(err)
	}

	return &dm
}

// CreateMapping creates the data schema for a specific collection.
func (s *Search) CreateMapping(documentMapping DocumentMapping) (bleve.Index, error) {
	// a generic reusable mapping for text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = false

	// a generic reusable mapping for datetime
	dateTimeFieldMapping := bleve.NewDateTimeFieldMapping()
	dateTimeFieldMapping.Store = false

	// a generic reusable mapping for number
	numericFieldMapping := bleve.NewNumericFieldMapping()
	numericFieldMapping.Store = false

	// a generic reusable mapping for boolean
	booleanFieldMapping := bleve.NewBooleanFieldMapping()
	booleanFieldMapping.Store = false

	// a generic reusable mapping for boolean
	geoPointFieldMapping := bleve.NewGeoPointFieldMapping()
	geoPointFieldMapping.Store = false

	collectionSchema := bleve.NewDocumentMapping()

	// iterate all the fields in the payload and create the field mappings for each of them - index name and mapping name is the same
	for fieldName, v := range documentMapping.Fields {
		if strings.HasPrefix(fieldName, "_") { // _ is for system only fields
			return nil, fmt.Errorf("field name: %s cannot start with _", fieldName)
		}

		fieldType := v.(map[string]interface{})

		switch fieldType["type"] {
		case "text":
			collectionSchema.AddFieldMappingsAt(fieldName, textFieldMapping)
		case "number":
			collectionSchema.AddFieldMappingsAt(fieldName, numericFieldMapping)
		case "datetime":
			collectionSchema.AddFieldMappingsAt(fieldName, dateTimeFieldMapping)
		case "boolean":
			collectionSchema.AddFieldMappingsAt(fieldName, booleanFieldMapping)
		case "geopoint":
			collectionSchema.AddFieldMappingsAt(fieldName, geoPointFieldMapping)
		default:
			log.Errorf("The data type: %s for field: %s is not valid.", fieldType["type"], fieldName)
			return nil, fmt.Errorf("the data type: %s for field: %s is not valid", fieldType["type"], fieldName)
		}
	}

	// System fields
	collectionSchema.AddFieldMappingsAt("_blockId", textFieldMapping)
	collectionSchema.AddFieldMappingsAt("_publicKey", textFieldMapping)
	collectionSchema.AddFieldMappingsAt("_timestamp", dateTimeFieldMapping)
	collectionSchema.AddFieldMappingsAt("_type", textFieldMapping)
	collectionSchema.AddFieldMappingsAt("_id", textFieldMapping) // transaction ID
	collectionSchema.AddFieldMappingsAt("_peerId", textFieldMapping)
	collectionSchema.AddFieldMappingsAt("_permittedAddresses", textFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping(documentMapping.Collection, collectionSchema)

	indexMapping.TypeField = "_type"
	indexMapping.DefaultAnalyzer = "en"
	indexMapping.StoreDynamic = false
	indexMapping.IndexDynamic = false

	collectionIndex, err := bleve.NewUsing(s.indexDirRoot+filepath.Dir("/")+documentMapping.Collection, indexMapping, scorch.Name, scorch.Name, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"method": "CreateMapping()",
		}).Error(err)
		return nil, err
	}

	err = s.db.Update(func(dbtx *bolt.Tx) error {
		collectionBucket, err := dbtx.CreateBucketIfNotExists([]byte(CollectionsBucket))

		if err != nil {
			return err
		}

		err = collectionBucket.Put([]byte(documentMapping.Collection), documentMapping.Serialize())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"method": "CreateMapping()",
		}).Error(err)
		return nil, err
	}

	collectionIndex.SetName(documentMapping.Collection) // rewrite the default name
	s.BlockchainIndices[documentMapping.Collection] = collectionIndex
	return collectionIndex, nil
}

// CreateMappingByJson creates the data schema for a specific collection, which is defined in JSON
// An example JSON payload:
// {
//     "collection": "new_collection",
//     "fields": {
//         "id": {"type": "text"},
//         "title": {"type": "text"},
//         "age": {"type": "number"},
//         "created": {"type": "datetime"},
//         "isModified": {"type": "boolean"},
//         "location": {"type": "geopoint"}
//     }
// }
func (s *Search) CreateMappingByJson(mappingJSON []byte) (bleve.Index, error) {
	var documentMapping DocumentMapping

	if err := json.Unmarshal(mappingJSON, &documentMapping); err != nil {
		log.Errorf("error parsing the document json mapping json payload: " + err.Error())
		return nil, err
	}

	if len(documentMapping.Collection) == 0 || documentMapping.Fields == nil {
		log.Errorf("%s is not a valid collection schema definition\n", mappingJSON)
		return nil, fmt.Errorf("%s is not a valid collection schema definition", mappingJSON)
	}

	if nil != s.BlockchainIndices[documentMapping.Collection] {
		log.Warnf("the collection " + documentMapping.Collection + " already exists. Nothing to do.")
		return nil, fmt.Errorf("the collection %s already exists. Nothing to do", documentMapping.Collection)
	}

	return s.CreateMapping(documentMapping)

}
