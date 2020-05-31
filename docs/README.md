# Blocace In 10 Minutes
This guide assumes you have an existing basic knowledge of Web API, database and digital signature. 

System prerequisites:

* (only when you prefer to compile Blocace server) Go version: 1.12 or later;

* (only when you prefer to compile Blocace server) GCC 5.1 or later. Windows may need to install [GCC](http://tdm-gcc.tdragon.net/download) if missing before installing the dependencies. Linux may also need to install gcc using the corresponding package management tool, like `yum install gcc` on RedHat or alike. macOS may need to install [Xcode Command Line Tools](https://www.ics.uci.edu/~pattis/common/handouts/macmingweclipse/allexperimental/macxcodecommandlinetools.html).
* [Node.js](https://nodejs.org/)

* JS libraries (install using `npm install` in [blocace-js](https://github.com/codingpeasant/blocace-js)):

```javascript
{
  "dependencies": {
    "ethers": "^4.0.43",
    "axios": "^0.19.1"
  }
}

```

## Step 1: Compile and start Blocace server

```bash
$ git clone https://github.com/codingpeasant/blocace.git
$ cd blocace
$ go get
$ go build -ldflags="-s -w -X main.version=0.1.0"
$ ./blocace server # it'd be blocace.exe for Windows

		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__  ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition 0.1.0

INFO[2020-05-31T15:38:27-04:00] configurations:                               advertiseAddress= bulkLoading=false hostP2p=0.0.0.0 loglevel=debug maxtime=2000 maxtx=2048 path=./data0 peerAddresses= portP2p=6091 porthttp=6899
INFO[2020-05-31T15:38:27-04:00] cannot find the db file. creating new...     
INFO[2020-05-31T15:38:27-04:00] cannot open index, path does not exist: ./data0/collections. creating the default collection instead... 
INFO[2020-05-31T15:38:27-04:00] the admin account has been created and registered successfully 

####################
PRIVATE KEY: 1471dc3e2fb9d2d43c2c598ff98f90c147c41a7f8899d490bee9e813ab5bcfe6
WARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.
####################

INFO[2020-05-31T15:38:27-04:00] did not find peer db dir ./data0/peers, creating one...
INFO[2020-05-31T15:38:27-04:00] no peer address(es) provided, starting without trying to discover
INFO[2020-05-31T15:38:28-04:00] begin to monitor transactions every 2000 milliseconds...
INFO[2020-05-31T15:38:28-04:00] awaiting signal...
```
By default, __Blocace__ creates a `data` directory within the working dir to store the blockchain and DB collections; the time interval to generate a block is 2 seconds; the max number of transactions (about documents) is 2048; it listens on port 6899 for web API calls; advertiseAddress bonds to loopback network interface; the bulk loading API is disabled; peerAddresses is empty (single node mode); the port open for P2P traffic is 6091. Please keep a note of the `root private key` which will be used to make administration API calls to Blocace server (step 2 below).

> If you are running this in a bash terminal on linux, you can quickly spin up a Blocace cluster by running `./start_cluster.sh 3` to compile and spin up 3 Blocace server nodes on localhost. To terminate the 3 nodes, simply use Ctrl + C in the terminal.

## Step 2: Run example.js with the root admin account private key
```bash
# open a new terminal tab and run
$ git clone https://github.com/codingpeasant/blocace-js.git
$ cd blocace-js
$ npm install
$ node ./example.js <PRIVATE KEY>

Private key decrypted: true

JWT (admin): eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlTmFtZSI6ImFkbWluIiwiYWRkcmVzcyI6IjB4MmIyOTNkOUIyM0JiNkZlRTM2OTY4OTlkNmMxQURjQWY1N0ZGQjM3RCIsImF1ZCI6ImJsb2NhY2UgdXNlciIsImV4cCI6MTU5MDk1NzExOSwiaWF0IjoxNTkwOTU2NTE5LCJpc3MiOiJibG9jYWNlIn0.TUiA4G1rp-E55_3cqyHbk01-lIk3bUsho7f9a-LZeR8

Address of new account: 0xf55486314B0C4F032d603B636327ed5c82218688

New account Info: {"address":"699 Canton Court, Mulino, South Dakota, 9647","collectionsReadOverride":null,"collectionsWrite":null,"dateOfBirth":"2018-10-01","email":"hoopervincent@mitroc.com","firstName":"Hooper","lastName":"Vincent","organization":"MITROC","phone":"+1 (849) 503-2756","position":"VP of Marketing","publicKey":"04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa","roleName":"user"}

Account permission response: {"message":"account permission updated","address":"0xf55486314B0C4F032d603B636327ed5c82218688"}

Get the update account: {"address":"699 Canton Court, Mulino, South Dakota, 9647","collectionsReadOverride":["default","collection2","new1"],"collectionsWrite":["default","new1"],"dateOfBirth":"2018-10-01","email":"hoopervincent@mitroc.com","firstName":"Hooper","lastName":"Vincent","organization":"MITROC","phone":"+1 (849) 503-2756","position":"VP of Marketing","publicKey":"04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa","roleName":"user"}
New collection info: {"message":"collection new1 created"}

JWT (new user): eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlTmFtZSI6InVzZXIiLCJhZGRyZXNzIjoiMHhmNTU0ODYzMTRCMEM0RjAzMmQ2MDNCNjM2MzI3ZWQ1YzgyMjE4Njg4IiwiYXVkIjoiYmxvY2FjZSB1c2VyIiwiZXhwIjoxNTkwOTU3MTE5LCJpYXQiOjE1OTA5NTY1MTksImlzcyI6ImJsb2NhY2UifQ.OH1XuB_ECCdwLRgMyp1DiDab6IzDBUX0zpd-0vznsxo

Update account response: {"message":"account updated","address":"0xf55486314B0C4F032d603B636327ed5c82218688"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"69292d123e8278e18e040fe7080898b4f6695413bd8890c851251b6646e4be82"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"8841661dc86c2fbc2586f3f658b72713e371d89efae562d848f0ef4329a78280"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"4da4d28f757484cb26ff94d94df6154d3676d33e00a0afd5dead650abe42c217"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"7494edfee13f844b71cea5735f7566c2e01cca3f3be8746dd43551fc1fb67d0b"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"a8af696e9eb5d84d5f504b190c7150e1ec1a0306c2453e1151937d9430dc18d9"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"052c7ea57c35174d075af7e67f02094fc9918bcc5d266250cc1e47cb48e6c859"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"8b389d537fd3fabb5a8a5fb00337ff30cfea6751f34d85321051be766ac7fd46"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"654998deca958eeef0c78e0b048529515a11fb18a73963d67dc2f2cf9baeaced"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"d441505afc8d491ba6cb79a6a3f2f28f3bcbc6e16bea14fafcddf373e3dba72b"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"a792c9491cb2abe5160d2c73723608bd3b9c87da6624c5b9d36611efd8909eb9"}

Waiting for the document to be included in a block... 

Query result: {"collection":"new1","status":{"total":1,"failed":0,"successful":1},"total_hits":10,"hits":[{"_id":"8b389d537fd3fabb5a8a5fb00337ff30cfea6751f34d85321051be766ac7fd46","_blockId":"373fb73d7d39436932e7cb969e97c42ae13886b6c1bf64c19d84479859a21192","_blockchainId":"99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410","_source":"{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}","_timestamp":"2020-05-31T16:22:00.001-04:00","_signature":"98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5","_address":"0xf55486314B0C4F032d603B636327ed5c82218688"},{"_id":"a8af696e9eb5d84d5f504b190c7150e1ec1a0306c2453e1151937d9430dc18d9","_blockId":"373fb73d7d39436932e7cb969e97c42ae13886b6c1bf64c19d84479859a21192","_blockchainId":"99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410","_source":"{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}","_timestamp":"2020-05-31T16:21:59.988-04:00","_signature":"98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5","_address":"0xf55486314B0C4F032d603B636327ed5c82218688"},{"_id":"654998deca958eeef0c78e0b048529515a11fb18a73963d67dc2f2cf9baeaced","_blockId":"373fb73d7d39436932e7cb969e97c42ae13886b6c1bf64c19d84479859a21192","_blockchainId":"99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410","_source":"{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}","_timestamp":"2020-05-31T16:22:00.011-04:00","_signature":"98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5","_address":"0xf55486314B0C4F032d603B636327ed5c82218688"}]}

Document included in the block: true

The document's integrity check passed: true

Block information: {"blockchainId":"99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410","blockId":"373fb73d7d39436932e7cb969e97c42ae13886b6c1bf64c19d84479859a21192","prevBlockId":"c39e703c750f99f5a774ba52615ff0b55a72edf516a687e6a1e8e5658e083a2b","blockHeight":1,"totalTransactions":10}

Blockchain information: [{"blockchainId":"99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410","tipBlockId":"373fb73d7d39436932e7cb969e97c42ae13886b6c1bf64c19d84479859a21192","lastHeight":1,"totalTransactions":11}]

All collections in the blockchain: {"message":"ok","collections":["default","new1"]}

Collection new1 data schema: {"message":"ok","mapping":{"collection":"new1","fields":{"age":{"encrypted":true,"type":"number"},"gender":{"type":"text"},"guid":{"type":"text"},"id":{"type":"text"},"isActive":{"type":"boolean"},"location":{"type":"geopoint"},"name":{"encrypted":true,"type":"text"},"registered":{"type":"datetime"},"tags":{"type":"text"}}}}

Peers of 99cd155750514bf8aeb2f87cdf25fab3877301e6cfdeb2ab769fe0a8c7283410: []
```
> Troubleshooting: if you see `{ message: 'public key not valid' }` error running `example.js` above, you can switch back to the Blocace server terminal, stop the server by Ctrl + C, remove the `./data*` folder(s), restart the server to generate a new private key.

That's it. You have successfully built Blocace server and accessed *ALL* its web APIs using the Blocace Javascript client.

# Step-by-step breakdown of example.js
> If you'd like to know more about the APIs, please continue reading.

## Setup root account
```javascript
var blocace = Blocace.createFromPrivateKey(process.argv[2])

// encrypt and decrypt the seed
var encryptPrivKey = blocace.encryptPrivateKey('123456')
var decryptPrivKey = Blocace.decryptPrivateKey(encryptPrivKey, '123456')

console.log('Private key decrypted: ' + (blocace.wallet.privateKey === decryptPrivKey) + '\n')
```
Output
```
Private key decrypted: true
```
## Create account using the root account (private key)
```javascript
// get JWT
const jwt = await blocace.getJWT()
console.log('JWT (admin): ' + jwt + '\n')

// register a new user account
const accountRes = await Blocace.createAccount(accountPayload, 'http', 'localhost', '6899')
console.log('Address of new account: ' + accountRes.data.address + '\n')

// get account
const account = await blocace.getAccount(accountRes.data.address)
console.log('New account Info: ' + JSON.stringify(account) + '\n')onsole.log(error);
});
```
Output
```
JWT (admin): eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlTmFtZSI6ImFkbWluIiwiYWRkcmVzcyI6IjB4RDE2MjFGNzZiMzMzOWIyRUFENTA2ODU5ZGRFRWRhRkZBMWYxOGM1MiIsImF1ZCI6ImJsb2NhY2UgdXNlciIsImV4cCI6MTU4MDM2MTAyOCwiaWF0IjoxNTgwMzYwNDI4LCJpc3MiOiJibG9jYWNlIn0.rKqkdaD-k8HmUW-z0W9WI41SUs7_sqSFdjGePdrYtKQ

Address of new account: 0xf55486314B0C4F032d603B636327ed5c82218688

New account Info: {"address":"699 Canton Court, Mulino, South Dakota, 9647","collectionsReadOverride":null,"collectionsWrite":null,"organization":"MITROC","dateOfBirth":"2018-10-01","email":"hoopervincent@mitroc.com","firstName":"Hooper","lastName":"Vincent","phone":"+1 (849) 503-2756","position":"VP of Marketing","publicKey":"04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa","roleName":"user"}
```
First login as the root admin user and obtain a [JSON Web Token](https://jwt.io/) to access Blocace server and create a new user account without read/write permissions. And then get the account information noting that `"collectionsReadOverride":null,"collectionsWrite":null`

## Grant collection-level permission to the new user
```javascript
// set the new account read / write permission
const accountPermissionRes = await blocace.setAccountReadWrite(permission, accountRes.data.address)
console.log('Account permission response: ' + JSON.stringify(accountPermissionRes.data) + '\n')

// get the user account again
const accountUpdated = await blocace.getAccount(accountRes.data.address)
console.log('Get the update account: ' + JSON.stringify(accountUpdated))
```
Output
```
Account permission response: {"message":"account permission updated","address":"0xf55486314B0C4F032d603B636327ed5c82218688"}

Get the update account: {"address":"699 Canton Court, Mulino, South Dakota, 9647","collectionsReadOverride":["default","collection2","new1"],"collectionsWrite":["default","new1"],"organization":"MITROC","dateOfBirth":"2018-10-01","email":"hoopervincent@mitroc.com","firstName":"Hooper","lastName":"Vincent","phone":"+1 (849) 503-2756","position":"VP of Marketing","publicKey":"04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa","roleName":"user"}
```
Setting the read/write permission to the new user: `'collectionsWrite': ['default', 'new1'],'collectionsReadOverride': ['default', 'collection2', 'new1']`.

## Create a new collection as the root admin user
```javascript
// create collection
const collectionCreationRes = await blocace.createCollection(collectionMappingPaylod)
console.log('New collection info: ' + JSON.stringify(collectionCreationRes) + '\n')

```
Output
```
New collection info: {"message":"collection new1 created"}
```
Create the collection (or table in SQL databases) with defined schema.

## Update the new user account information
```javascript
// initialize the new user account
var blocaceUser = Blocace.createFromPrivateKey('277d271593d205c6078964c31fb393303efd76d5297906f60d2a7a7d7d12c99a')
// get JWT for the user account
const jwtUser = await blocaceUser.getJWT()
console.log('JWT (new user): ' + jwtUser + '\n')

// update account
accountPayload.email = 'asd@asd.com'
const accountUpdateRes = await blocaceUser.updateAccount(accountPayload, accountRes.data.address)
console.log('Update account response: ' + JSON.stringify(accountUpdateRes.data) + '\n')
```
Output
```
JWT (new user): eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlTmFtZSI6InVzZXIiLCJhZGRyZXNzIjoiMHhmNTU0ODYzMTRCMEM0RjAzMmQ2MDNCNjM2MzI3ZWQ1YzgyMjE4Njg4IiwiYXVkIjoiYmxvY2FjZSB1c2VyIiwiZXhwIjoxNTgwMzYxMDI4LCJpYXQiOjE1ODAzNjA0MjgsImlzcyI6ImJsb2NhY2UifQ.UBw-D7AL1KNBl-Ww2NHz-HvV92BNrfcmdXyb0HwzjGI

Update account response: {"message":"account updated","address":"0xf55486314B0C4F032d603B636327ed5c82218688"}
```
Login as the new user and update its own account's email address. Note that the account information is just for convenience to track identity and doesn't affect the usage of Blocace

## Sign and put documents to Blocace and query them
```javascript
// put 10 documents
for (let index = 0; index < 10; index++) {
  const putDocRes = await blocaceUser.signAndPutDocument(document, 'new1')
  console.log('Put document response: ' + JSON.stringify(putDocRes) + '\n')
}

// wait for block to be generated
await timeout(2000)
console.log('Waiting for the document to be included in a block... \n')

// query the blockchain
const queryRes = await blocaceUser.query(queryPayload, 'new1')
console.log('Query result: ' + JSON.stringify(queryRes) + '\n')
```
Output
```
Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"8a545086ebfac8d7f38c08ceb618f2afe35850e9ba9890784abe89288f42e7bd"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"dd6182df7f97a8df1bcbfe9c107e369a002b03a62114f5f7152460ad98194e03"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"01fa1686554b585e1436436c2cff40bb7b250eb383699dcebd389ff4af504e50"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"e4644e6d64fdc2f45526742e4921010f48b29ae8fe1b8655ef544853b7acd10c"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"28728359c7e240dddcb83e15e7f078ba45f329b1202cd0ca0ad9d11ba4945814"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"98c41ffa3227d8f8674a3b8865d7d0e4622815e3895acc6e0da8d1b8caf39084"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"516ab6ec7db085b0347b7a5f67b36e6654092bc60cc40b2ec3e6370999ef42a3"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"8c62c28482b098eba844471289f2ddfad1e1a6748c389d8348e96df017841b33"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"f8dde1543a7d644fc1ec6e1765c0e694fc96f51625c4d83926b611959188739d"}

Put document response: {"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"07245c7a01cf7fac705a40b3e1632bcc06754e6ce7f5d01624d66e9b567d91ca"}

Waiting for the document to be included in a block... 

Query result: 
{
	"collection": "new1",
	"status": {
		"total": 1,
		"failed": 0,
		"successful": 1
	},
	"total_hits": 10,
	"hits": [{
		"_id": "8a545086ebfac8d7f38c08ceb618f2afe35850e9ba9890784abe89288f42e7bd",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.624-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}, {
		"_id": "f8dde1543a7d644fc1ec6e1765c0e694fc96f51625c4d83926b611959188739d",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.712-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}, {
		"_id": "516ab6ec7db085b0347b7a5f67b36e6654092bc60cc40b2ec3e6370999ef42a3",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.691-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}]
}
```
Sign each of the documents with the new user's private key and send them to Blocace. Blocace server verifies the digital signature of each document and put them to the transaction queue. The goroutine to generate blocks dequeues transactions periodically and append them to the blockchain so all the transactions in each block are indexed to be queried later. The query we use in this example is:
```javascript
{
  'size': 3,
  'from': 0,
  'query': {
    'match': 'Compton',
    'field': 'name'
  }
}
```

## Verify the integrity of the documents
```javascript
// verify if the transaction is included in the block (by block merkle tree rebuild)
const verificationPassed = await blocaceUser.verifyTransaction(queryRes.hits[0]._blockchainId, queryRes.hits[0]._blockId, queryRes.hits[0]._id)
console.log('Document included in the block: ' + verificationPassed + '\n')

// verify signature
console.log('The document\'s integrity check passed: ' + Blocace.verifySignature(queryRes.hits[0]._source, queryRes.hits[0]._signature, blocaceUser.wallet.address) + '\n')
```
Output
```
Document included in the block: true

The document's integrity check passed: true
```
The Blocace client first verifies that the document has been persisted in the blockchain and that the document is not tempered with. This is the blockchain philosophy:
> Don't Trust. Verify!

## Get block, blockchain and collection administration information
```javascript
// get block information
const blockRes = await blocace.getBlockInfo(queryRes.hits[0]._blockId)
console.log('Block information: ' + JSON.stringify(blockRes) + '\n')

// get blockchain information
const blockchainRes = await blocace.getBlockchainInfo()
console.log('Blockchain information: ' + JSON.stringify(blockchainRes) + '\n')

// get all collections
const collectionsRes = await blocace.getCollections()
console.log('All collections in the blockchain: ' + JSON.stringify(collectionsRes) + '\n')

// get collection data schema
const collectionRes = await blocace.getCollection('new1')
console.log('Collection new1 data schema: ' + JSON.stringify(collectionRes) + '\n')
```
Output
```
Block information: {"blockId":"cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8","lastBlockId":"47e7023f02c4f762d458e674ce1075666e47cafa93a701b6cb88615c6b4f6dc5","blockHeight":1,"totalTransactions":10}

Blockchain information: {"newestBlockId":"cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8","lastHeight":1,"totalTransactions":11}

All collections in the blockchain: {"message":"ok","collections":["default","new1"]}

Collection new1 data schema: {"message":"ok","mapping":{"collection":"new1","fields":{"age":{"encrypted":true,"type":"number"},"gender":{"type":"text"},"guid":{"type":"text"},"id":{"type":"text"},"isActive":{"type":"boolean"},"location":{"type":"geopoint"},"name":{"encrypted":true,"type":"text"},"registered":{"type":"datetime"},"tags":{"type":"text"}}}}
```
Blocace client is able to get adminstration information about a given block. In this example, the `blockHeight` is `1` as this block is the 2nd in the blockchain (after the genesis block). It has `10` transaction documents that we just put; the blockchain has `11` transaction documents: 1 `genesis transactions` + 10 `user transactions`; it also gets the schema of collection `new1`.

## Get peer(s) of the current Blocace node
```javascript
// get peer nodes of a given node
const peerRes = await blocace.getPeers()
console.log('Peers of ' + queryRes.hits[0]._blockchainId + ': ' + JSON.stringify(peerRes) + '\n')
```
Output
```
Peers of cba1739c9aeabc0698a5da2a701f22c6f2e4c0122b2a44d019bd4b6105a4801c: [{"public_key":"45b3c9f26eef3b0bfbdea99af1731ab7a08a1dfcaf58a7e8abf55db73eda45c9","address":"::","Port":6092,"Address":":6092"},{"public_key":"70a50de06cbc9e36010e9277c6ae74bafce8af4e6c4b168b864a373a3a8d1a1f","address":"::","Port":6093,"Address":":6093"}]
```
In this example, the blocace-js client is configured talking to a single server. The `getPeers()` returns the alive peer(s) that are currently known this server.

> You're all set! Go ahead an build your web DAPP around Blocace!

# Usage Reference
Note that the APIs are constantly changing in the pre-release phase.
## Blocace CLI reference
### Server CLI
The major command to start a Blocace instance.
```
$ ./blocace s -h
		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__  ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition 0.1.0

NAME:
   blocace server - start the major blocace server

USAGE:
   blocace server [command options] [arguments...]

OPTIONS:
   --dir value, -d value               the path to the folder of data persistency (default: "data")
   --secret value, -s value            the password to encrypt data and manage JWT
   --maxtx value, -m value             the max transactions in a block (default: 2048)
   --maxtime value, -t value           the time in milliseconds interval to generate a block (default: 2000)
   --porthttp value, -o value          the port that the web api http server listens on (default: "6899")
   --portP2p value, -p value           the port that the p2p node listens on (default: 6091)
   --hostP2p value, -w value           the hostname/ip address that the p2p node binds to (default: "0.0.0.0")
   --advertiseAddress value, -a value  the public address of this node which is advertised on the ID sent to peers during a handshake protocol (optional)
   --peerAddresses value, -e value     the comma-separated address:port list of the peers (optional)
   --bulkLoading value, -b value       enable bulking loading API (default: "false")
   --loglevel value, -l value          the log levels: panic, fatal, error, warn, info, debug, trace (default: "info")

```
Example:
```
$ ./blocace s -l debug
		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__  ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition 0.1.0

INFO[2020-05-31T15:38:27-04:00] configurations:                               advertiseAddress= bulkLoading=false hostP2p=0.0.0.0 loglevel=debug maxtime=2000 maxtx=2048 path=./data0 peerAddresses= portP2p=6091 porthttp=6899
INFO[2020-05-31T15:38:27-04:00] cannot find the db file. creating new...     
INFO[2020-05-31T15:38:27-04:00] cannot open index, path does not exist: ./data0/collections. creating the default collection instead... 
INFO[2020-05-31T15:38:27-04:00] the admin account has been created and registered successfully 

####################
PRIVATE KEY: 1471dc3e2fb9d2d43c2c598ff98f90c147c41a7f8899d490bee9e813ab5bcfe6
WARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.
####################

INFO[2020-05-31T15:38:27-04:00] did not find peer db dir ./data0/peers, creating one... 
INFO[2020-05-31T15:38:27-04:00] no peer address(es) provided, starting without trying to discover 
INFO[2020-05-31T15:38:28-04:00] begin to monitor transactions every 2000 milliseconds... 
INFO[2020-05-31T15:38:28-04:00] awaiting signal... 
```
### Key generation CLI
In case the Blocace administrator lost the root admin account, this command recreates it.
```
$ ./blocace k -h

		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__ ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition 0.1.0

NAME:
   blocace keygen - generate and register an admin account

USAGE:
   blocace keygen [command options] [arguments...]

OPTIONS:
   --dir value, -d value  the path to the folder of data persistency (default: "data")
```
Example:
```
$ ./blocace k

		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__ ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition 0.1.0

INFO[2020-02-01T12:02:30-05:00] db file exists. generating an admin keypair and registering an account... 
INFO[2020-02-01T12:02:30-05:00] the account has been created and registered successfully 

####################
PRIVATE KEY: 81244df62f43a163a2f4a4894ef531ba1a493b921fb3bbaabdb2222e632f7734
WARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.
####################
```
## Blocace web API reference
### `static create(protocol, hostname, port)`
Generate random Blocace client key pair and initialize the client class

Example:
```
var blocace = Blocace.create('http', 'localhost', '6899')
```
### `static createFromPrivateKey(privKey, protocol, hostname, port)`
Use an existing client private key and initialize the client class

Example:
```
var blocace = Blocace.createFromPrivateKey('81244df62f43a163a2f4a4894ef531ba1a493b921fb3bbaabdb2222e632f7734)
```

### `encryptPrivateKey(password)`
Get the encrypted private key. The return value is a concatenation of the salt, IV and the cipher text of the private key

Example:
```
var encryptPrivKey = blocace.encryptPrivateKey('123456')
```

### `static decryptPrivateKey(encrypted, password)`
Decrypt the private key from the encryption string, which is a concatenation of the salt, IV and the cipher text of the private key

Example:
```
var decryptPrivKey = Blocace.decryptPrivateKey(encryptPrivKey, '123456')
```

### `static verifySignature(rawDocument, signature, address)`
Verify if the signature of a document matches the claimed address (aka. public key). This API can be used to verify the integrity of a document

Example:
```
var isValidSignature = Blocace.verifySignature(queryRes.hits[0]._source, queryRes.hits[0]._signature, blocaceUser.wallet.address)
```

### `getPublicKey()`
Get public key of the wallet

Example:
```
var publicKey = blocace.getPublicKey()
```

### `static async createAccount(accountPayload)`
Create a new account

Example:
```
const accountPayload = {
  'dateOfBirth': '2018-10-01',
  'firstName': 'Hooper',
  'lastName': 'Vincent',
  'organization': 'MITROC',
  'position': 'VP of Marketing',
  'email': 'hoopervincent@mitroc.com',
  'phone': '+1 (849) 503-2756',
  'address': '699 Canton Court, Mulino, South Dakota, 9647',
  'publicKey': 'b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa'
}

const accountRes = await Blocace.createAccount(accountPayload, 'http', 'localhost', '6899')
```

### `async updateAccount(accountPayload, address)`
Update the current account

Example:
```
const accountPayload = {
  'dateOfBirth': '2018-10-01',
  'firstName': 'Hooper',
  'lastName': 'Vincent',
  'organization': 'MITROC',
  'position': 'VP of Marketing',
  'email': 'hoopervincent@mitroc.com',
  'phone': '+1 (849) 503-2756',
  'address': '699 Canton Court, Mulino, South Dakota, 9647',
  'publicKey': 'b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa'
}

accountPayload.email = 'asd@asd.com'
const accountUpdateRes = await blocaceUser.updateAccount(accountPayload, accountRes.data.address)
```
Output:
```
{"address":"699 Canton Court, Mulino, South Dakota, 9647","collectionsReadOverride":null,"collectionsWrite":null,"organization":"MITROC","dateOfBirth":"2018-10-01","email":"hoopervincent@mitroc.com","firstName":"Hooper","lastName":"Vincent","phone":"+1 (849) 503-2756","position":"VP of Marketing","publicKey":"04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa","roleName":"user"}
```
### `async setAccountReadWrite(permissionPayload, address)`
Grand collection level read/write permission

Example:
```
const accountPermissionRes = await blocace.setAccountReadWrite(permission, accountRes.data.address)
```
Output:
```
{"message":"account permission updated","address":"0xf55486314B0C4F032d603B636327ed5c82218688"}
```
### `async getChallenge()`
A challenge issued from Blocace server for the client to authenticate

Example:
```
const challengeResponse = await this.getChallenge()
```
### `async getJWT()`
Get the challenge, give back the solution and obtain the JWT ([JSON Web Token](https://jwt.io/))

Example:
```
const jwt = await blocace.getJWT()
```
Output:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlTmFtZSI6ImFkbWluIiwiYWRkcmVzcyI6IjB4RDE2MjFGNzZiMzMzOWIyRUFENTA2ODU5ZGRFRWRhRkZBMWYxOGM1MiIsImF1ZCI6ImJsb2NhY2UgdXNlciIsImV4cCI6MTU4MDM2MTAyOCwiaWF0IjoxNTgwMzYwNDI4LCJpc3MiOiJibG9jYWNlIn0.rKqkdaD-k8HmUW-z0W9WI41SUs7_sqSFdjGePdrYtKQ
```
### `async getAccount(address)`
Get the account's information

Example:
```
const account = await blocace.getAccount(accountRes.data.address)
```
Output:
```
{
	"address": "699 Canton Court, Mulino, South Dakota, 9647",
	"collectionsReadOverride": null,
	"collectionsWrite": null,
	"organization": "MITROC",
	"dateOfBirth": "2018-10-01",
	"email": "hoopervincent@mitroc.com",
	"firstName": "Hooper",
	"lastName": "Vincent",
	"phone": "+1 (849) 503-2756",
	"position": "VP of Marketing",
	"publicKey": "04b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa",
	"roleName": "user"
}
```
### `async createCollection(collectionPayload)`
Create an new collection with schema

Example:
```
const collectionCreationRes = await blocace.createCollection(collectionMappingPaylod)
```
Output:
```
{"message":"collection new1 created"}
```
### `async signAndPutDocument(document, collection)`
Write and digitally sign a JSON document to add to a collection

Example:
```
const document = {
  'id': '5bf1d3fdf6fd4a5c4638f64e',
  'guid': 'f51b68c5-f274-4ce1-984f-b4fb4d618ff3',
  'isActive': false,
  'age': 28,
  'name': 'Carly Compton',
  'gender': 'male',
  'registered': '2015-09-18T12:59:51Z',
  'location': {
    'lon': 46.564666,
    'lat': 53.15213
  },
  'tags': [
    'incididunt',
    'dolore'
  ],
  'friends': [
    {
      'id': 0,
      'name': 'Jimenez Byers'
    },
    {
      'id': 1,
      'name': 'Gabriela Mayer'
    }
  ],
  'notExist': 'haha'
}

const putDocRes = await blocaceUser.signAndPutDocument(document, 'new1')
```
Output:
```
{"status":"ok","fieldErrors":null,"isValidSignature":true,"transactionID":"8a545086ebfac8d7f38c08ceb618f2afe35850e9ba9890784abe89288f42e7bd"}
```
### `async putDocumentBulk(documents, collection)`
Write a bulk of JSON documents in a single HTTP request to a collection. WARNING: this makes the documents unverifiable

Example:
```
const payload = [
  {...},
  {...},
  {...}
]
await blocaceUser.putDocumentBulk(payload, 'new2')
```

### `async query(queryPayload, collection)`
Query the documents from Blocase with a query against a collection. Check out [Blevesearch Query](https://blevesearch.com/docs/Query/) for the query DSL.

Example:
```
const queryPayload = {
  'size': 3,
  'from': 0,
  'query': {
    'match': 'Compton',
    'field': 'name'
  }
}
const queryRes = await blocaceUser.query(queryPayload, 'new1')
```
Output:
```
{
	"collection": "new1",
	"status": {
		"total": 1,
		"failed": 0,
		"successful": 1
	},
	"total_hits": 10,
	"hits": [{
		"_id": "8a545086ebfac8d7f38c08ceb618f2afe35850e9ba9890784abe89288f42e7bd",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.624-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}, {
		"_id": "f8dde1543a7d644fc1ec6e1765c0e694fc96f51625c4d83926b611959188739d",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.712-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}, {
		"_id": "516ab6ec7db085b0347b7a5f67b36e6654092bc60cc40b2ec3e6370999ef42a3",
		"_blockId": "cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8",
		"_source": "{\"id\":\"5bf1d3fdf6fd4a5c4638f64e\",\"guid\":\"f51b68c5-f274-4ce1-984f-b4fb4d618ff3\",\"isActive\":false,\"age\":28,\"name\":\"Carly Compton\",\"gender\":\"male\",\"registered\":\"2015-09-18T12:59:51Z\",\"location\":{\"lon\":46.564666,\"lat\":53.15213},\"tags\":[\"incididunt\",\"dolore\"],\"friends\":[{\"id\":0,\"name\":\"Jimenez Byers\"},{\"id\":1,\"name\":\"Gabriela Mayer\"}],\"notExist\":\"haha\"}",
		"_timestamp": "2020-01-30T00:00:28.691-05:00",
		"_signature": "98c21b760b61fd4a59af9ea511f75f0338a76881bbd820ed3bb5c14a7dcf3d9847025cdf3aca07e7b448d8a7358d8678298afba8b3d9b16b9bac635457dccde5",
		"_address": "0xf55486314B0C4F032d603B636327ed5c82218688"
	}]
}
```
### `async verifyTransaction(blockchainId, blockId, transationId)`
Obtain a copy of block [Merkle Tree](https://en.wikipedia.org/wiki/Merkle_tree) and verify if the target document adding transaction has been included in the blockchain

Example:
```
const verificationPassed = await blocaceUser.verifyTransaction(queryRes.hits[0]._blockchainId, queryRes.hits[0]._blockId, queryRes.hits[0]._id)
```

### `async getBlockInfo(blockchainId, blockId)`
Get the information of a target block

Example:
```
const blockRes = await blocace.getBlockInfo(queryRes.hits[0]._blockchainId, queryRes.hits[0]._blockId)
```
Output:
```
{"blockId":"cfc01dc667753185a5635b33ebbff42b452476f15a4f63fceb210aad68dac3b8","lastBlockId":"47e7023f02c4f762d458e674ce1075666e47cafa93a701b6cb88615c6b4f6dc5","blockHeight":1,"totalTransactions":10}
```
### `async getBlockchainInfo()`
Get the information of all the chains in the network

Example:
```
const blockchainRes = await blocace.getBlockchainInfo()
```
Output:
```
[
    {
        "blockchainId": "bcfe7d257e28f656937b16b93448ebae57107e1536730c61600655a640194037",
        "tipBlockId": "86ccbffe9aac977eead58ff60d65d3d79022e6e191caf6bd3deffa5a1fffee4b",
        "lastHeight": 1,
        "totalTransactions": 11
    },
    {
        "blockchainId": "3cb1b5a4e3d6a5482e83381d747104e4e8ad6c78d7202760764079cc82b66bca",
        "tipBlockId": "b82ebc9e8e9eda31606c37c8772dd4c8b42c68579c149a34a0676dadce200ca4",
        "lastHeight": 5,
        "totalTransactions": 3001
    },
    {
        "blockchainId": "53388a5926ccc8ec4a6a129ea57fe0b86ffee6a1c033e6aca820e26d0e0c418c",
        "tipBlockId": "84a558c5a59b16711adc91cd7716775756c39a7475d0a06d1aedf89574eba9cf",
        "lastHeight": 0,
        "totalTransactions": 1
    }
]
```
### `async getPeers()`
Get the basic information of the alive peers known to node that the client currently talks to

Example:
```
const peerRes = await blocace.getPeers()
```
Output:
```
[
    {
        "public_key": "3cb1b5a4e3d6a5482e83381d747104e4e8ad6c78d7202760764079cc82b66bca",
        "address": "::",
        "Port": 16091,
        "Address": ":16091"
    },
    {
        "public_key": "53388a5926ccc8ec4a6a129ea57fe0b86ffee6a1c033e6aca820e26d0e0c418c",
        "address": "::",
        "Port": 26091,
        "Address": ":26091"
    }
]
```
### `async getCollections()`
Get all the collections in the blockchain

Example:
```
const collectionsRes = await blocace.getCollections()
```
Output:
```
{"message":"ok","collections":["default","new1"]}
```
### `async getCollection(collectionName)`
Get the information of a certain collection

Example:
```
const collectionRes = await blocace.getCollection('new1')
```
Output:
```
{
	"message": "ok",
	"mapping": {
		"collection": "new1",
		"fields": {
			"age": {
				"encrypted": true,
				"type": "number"
			},
			"gender": {
				"type": "text"
			},
			"guid": {
				"type": "text"
			},
			"id": {
				"type": "text"
			},
			"isActive": {
				"type": "boolean"
			},
			"location": {
				"type": "geopoint"
			},
			"name": {
				"encrypted": true,
				"type": "text"
			},
			"registered": {
				"type": "datetime"
			},
			"tags": {
				"type": "text"
			}
		}
	}
}
```

> Check out [example.js](https://github.com/codingpeasant/blocace-js/blob/master/example.js) for the full usage of the client lib.
