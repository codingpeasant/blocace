# Blocace Javascript Client
[![js-standard-style](https://cdn.rawgit.com/feross/standard/master/badge.svg)](https://github.com/feross/standard)

## The current available API are:
* `static create(hostname, port, protocol)` - Generate random Blocase client key pair and initialize the client
* `static createFromPrivateKey(privKey, hostname, port, protocol)` - Use an existing client private key and initialize the client
* `encryptPrivateKey(password)` - Get the encrypted private key
* `static decryptPrivateKey(encrypted, password)` - Decrypt the private key cipher text
* `static verifySignature(rawDocument, signature, address)` - Verify if the signature of a document matches the claimed address (aka. public key). This API can be used to verify the integrity of a document
* `async createAccount(accountPayload)` - Create a new account
* `async updateAccount(accountPayload, address)` - Update the account
* `async setAccountReadWrite(permissionPayload, address)` - Grand collection level read/write permission
* `async getChallenge()` - A challenge issued from Blocase server for the client to authenticate
* `async getJWT()` - Get the challenge, give back the solution and obtain the JWT ([JSON Web Token](https://jwt.io/))
* `async getAccount(address)` - Get the account's information
* `async createCollection(collectionPayload)` - Create an new collection with schema
* `async signAndPutDocument(document, collection)` - Write and digitally sign a JSON document to add to a collection
* `async putDocumentBulk(documents, collection)` - Write a bulk of JSON documents in a single HTTP request to a collection. WARNING: this makes the documents unverifiable
* `async query(queryPayload, collection)` - Query the documents from Blocase with a query against a collection
* `async verifyTransaction(blockId, transationId)` - Obtain a copy of block [Merkle Tree](https://en.wikipedia.org/wiki/Merkle_tree) and verify if the target document adding transaction has been included in the blockchain
* `async getBlockInfo(blockId)` - Get the information of a target block
* `async getBlockchainInfo()` - Get the information of the whole blockchain
* `async getCollections()` - Get all the collections in the blockchain
* `async getCollection(collectionName)` - Get the information of a certain collection

> Check out [example.js](https://github.com/codingpeasant/blocace/blob/master/client/example.js) for the full usage of the client lib.