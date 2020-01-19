# Blocace In 10 Minutes
> This guide assumes you have an existing basic knowledge of Web API, database and digital signature. 
> System prerequisites:
> * [Node.js](https://nodejs.org/)
> * JS libraries:

```javascript
{
  "dependencies": {
    "ethers": "^4.0.43",
    "axios": "^0.19.1"
  }
}

```

## Start Blocace server

```bash
$ ./blocace server

                 ____  __     __    ___   __    ___  ____
                (  _ \(  )   /  \  / __) / _\  / __)(  __)
                 ) _ (/ (_/\(  O )( (__ /    \( (__ ) _)
                (____/\____/ \__/  \___)\_/\_/ \___)(____)

                        Community Edition 0.0.1

time="2020-01-18T21:52:56-05:00" level=info msg="configurations: " maxtime=2000 maxtx=256 path=data port=6899
time="2020-01-18T21:52:56-05:00" level=info msg="cannot find the db file. creating new..."
time="2020-01-18T21:52:56-05:00" level=info msg="cannot (data\collections): cannot open index, path does not exist. creating the default collection instead..."
time="2020-01-18T21:52:56-05:00" level=info msg="the account has been created and registered successfully"

####################
PRIVATE KEY: a6df359954745422941e16b144594c704a74d591129981745efbf78e99ae53e0
WARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.
####################

time="2020-01-18T21:52:56-05:00" level=info msg="begin to monitor transactions every 2000 milliseconds..."
time="2020-01-18T21:52:56-05:00" level=info msg="awaiting signal..."

```
By default, __Blocace__ creates a `data` directory within the working dir to store the blockchain and DB collections; the time interval to generate a block is 2 seconds; the max number of transactions (about documents) is 256; it listens on port 6899 for web API calls.

## The following is a breakdown of Blocace REST APIs
> You can also skip reading this document for now and run [example.js](https://github.com/codingpeasant/blocace/blob/master/client/example.js) leveraging [Blocase JS client](https://github.com/codingpeasant/blocace/blob/master/client/index.js) directly to quickly get your hands dirty.

### Create account
```javascript
"use strict"

const ethUtil = require("ethereumjs-util")
const ethHDKey = require('ethereumjs-wallet/hdkey')
const axios = require('axios')

// The master seed is a mnemonic code or mnemonic sentence -- a group of easy to remember words -- for the generation of deterministic wallets.
const HDKey = ethHDKey.fromMasterSeed("guess tortoise flavor sorry brand ten faculty assist green reopen best gaze")
const wallet = HDKey.getWallet()

axios.post('http://localhost:6899/account', {
	"dateOfBirth": "2018-10-01",
	"firstName": "Hooper",
	"lastName": "Vincent",
	"company": "MITROC",
	"email": "hoopervincent@mitroc.com",
	"phone": "+1 (849) 503-2756",
	"address": "699 Canton Court, Mulino, South Dakota, 9647",
	"publicKey": wallet.getPublicKey().toString("hex")
}).then(function (response) {
    console.log(response.data);
}).catch(function (error) {
    console.log(error);
});

```

`/account` response
```javascript
{
    "message": "account created",
    "address": "0x730C5da9d5A7B39AD3B2e2274525B5eb2A9fa28D"
}
```

### Generate challenge and get JWT
__Blocace__ web API needs authentication/authorization to access. [JSON Web Token](https://jwt.io) is used for this purpose.

```javascript
const getChallenge = async () => {
    try {
      return axios.get('http://localhost:6899/jwt/challenge/0x730C5da9d5A7B39AD3B2e2274525B5eb2A9fa28D')
    } catch (error) {
      console.error(error)
    }
}

const getJWT = async () => {
    const challengeResponse = await getChallenge()
    var challengeHash = ethUtil.keccak(challengeResponse.data.challenge, 256)
    var sig = ethUtil.ecsign(challengeHash, Buffer.from(wallet.getPrivateKey(), 'hex'))
    axios.post('http://localhost:6899/jwt', {
        "address": "0x730C5da9d5A7B39AD3B2e2274525B5eb2A9fa28D",
        "signature": sig.r.toString("hex") + sig.s.toString("hex")
    }).then(response => {
        if (response.data.token) {
          console.log(
            `Token: ${response.data.token}`
          )
        }
    }).catch(error => {
        console.log(error)
    })
}

getJWT()
```

`/jwt/challenge` response:
```javascript
{
    "message": "challenge word created",
    "challenge": "f6HlouQByP9VKJCfFMLwFUbJh45YCevmuoadflSQNfZGqrahkdrJ2j5dVDALckjd"
}
```

`/jwt` response:
```javascript
{
	"message": "JWT issued",
	"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzEzMDgyLCJpYXQiOjE1NDYzMTI0ODIsImlzcyI6ImJsb2Nhc2UifQ.EG0pM1dNOU3V4W2cKwePflWzNMooTG3saOtGBb_2rCE"
}
```

### Get account

From this step on, we have to use the token from `/jwt` to access the web APIs. The token expires in 10 minutes by default.

```javascript
axios.request({
    url: 'http://localhost:6899/account/0x730C5da9d5A7B39AD3B2e2274525B5eb2A9fa28D',
    method: 'get',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzEzNDEzLCJpYXQiOjE1NDYzMTI4MTMsImlzcyI6ImJsb2Nhc2UifQ.ST3iT_TiIp0gKrge1NIw7kgtDK_Tic8JXhDWrYuC0Yo'}
}).then(response => {
    if (response.data) {
        console.log(
        `Account: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/account` response:
```javascript
{
	"address": "699 Canton Court, Mulino, South Dakota, 9647",
	"company": "MITROC",
	"dateOfBirth": "2018-10-01",
	"email": "hoopervincent@mitroc.com",
	"firstName": "Hooper",
	"lastName": "Vincent",
	"phone": "+1 (849) 503-2756"
}
```

### Create collection

`/collection` creates the data schema of the collections. Valid data typtes are: `text`, `number`, `datetime` ([RFC3339 format](https://www.ietf.org/rfc/rfc3339.txt)), `boolean`, `geopoint`. It also supports data encryption on fields.
```javascript
axios.request({
    url: 'http://localhost:6899/collection',
    method: 'post',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE0NDMwLCJpYXQiOjE1NDYzMTM4MzAsImlzcyI6ImJsb2Nhc2UifQ.Gu2VwItrrBuB64foM8PSGjJlFmxEjwVM3pCujStkL44'},
    data: {
        "collection": "new",
        "fields": {
            "id": {"type": "text"},
            "guid": {"type": "text"},
            "age": {"type": "number", "encrypted": true},
            "registered": {"type": "datetime"},
            "isActive": {"type": "boolean"},
            "gender": {"type": "text"},
            "name": {"type": "text","encrypted": true},
            "location": {"type": "geopoint"},
            "tags": {"type": "text"}
        }
    }
}).then(response => {
    if (response.data) {
        console.log(
        `Response: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/collection` response:
```javascript
{
	"message": "collection new created"
}
```

### Put a JSON document to the collection

```javascript
const document = {
    "id": "5bf1d3fdf6fd4a5c4638f64e",
    "guid": "f51b68c5-f274-4ce1-984f-b4fb4d618ff3",
    "isActive": false,
    "age": 28,
    "name": "Carly Compton",
    "gender": "male",
    "registered": "2015-09-18T12:59:51Z",
    "location": {
        "lon": 46.564666,
        "lat": 53.15213
    },
    "tags": [
        "incididunt",
        "dolore"
    ],
    "friends": [
        {
            "id": 0,
            "name": "Jimenez Byers"
        },
        {
            "id": 1,
            "name": "Gabriela Mayer"
        }
    ],
    "notExist": "haha"
}

var docHash = ethUtil.keccak(JSON.stringify(document), 256)
var sig = ethUtil.ecsign(docHash, Buffer.from(wallet.getPrivateKey(), 'hex'))

axios.request({
    url: 'http://localhost:6899/document/new',
    method: 'post',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE1MjAyLCJpYXQiOjE1NDYzMTQ2MDIsImlzcyI6ImJsb2Nhc2UifQ.zqFbNycKJfH-LJhfaYVyGcszMbmA6-aZ5l9yyXsCws8'},
    data: {
        "rawDocument": JSON.stringify(document),
        "signature": sig.r.toString("hex") + sig.s.toString("hex")
    }
}).then(response => {
    if (response.data) {
        console.log(
        `Response: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/document/{collection}` response:
```javascript
{
	"status": "ok",
	"fieldErrors": null,
	"isValidSignature": true,
	"transactionID": "82a9c23396dcf047f01cca6923541350a1b90fd37274c63a881a19ba1e97e0da"
}
```

### Query documents
__Blocace__ is shipped with a complete query DSL. The following example is to find all the documents containing `Carly` in field `name`.

```javascript
axios.request({
    url: 'http://localhost:6899/search/new',
    method: 'post',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE1OTk1LCJpYXQiOjE1NDYzMTUzOTUsImlzcyI6ImJsb2Nhc2UifQ.Mc6354XI4VSNi8le-v4Fce5M3PYTf0sbqvDA7jNJ-AI'},
    data: {
        "size": 10,
        "from": 0,
        "query": {
            "match": "Carly",
            "field": "name"
        }
    }
}).then(response => {
    if (response.data) {
        console.log(
        `Response: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/search/{collection}` response:
```javascript
{
	"collection": "new",
	"status": {
		"total": 1,
		"failed": 0,
		"successful": 1
	},
	"total_hits": 2,
	"hits": [{
		"_id": "82a9c23396dcf047f01cca6923541350a1b90fd37274c63a881a19ba1e97e0da",
		"_blockId": "33145e8dfc65b7a6a48f597e13574a32da939942c04f7c5b167b27417e13cb46",
		"_source": {
			"age": 28,
			"friends": [{
				"id": 0,
				"name": "Jimenez Byers"
			}, {
				"id": 1,
				"name": "Gabriela Mayer"
			}],
			"gender": "male",
			"guid": "f51b68c5-f274-4ce1-984f-b4fb4d618ff3",
			"id": "5bf1d3fdf6fd4a5c4638f64e",
			"isActive": false,
			"location": {
				"lat": 53.15213,
				"lon": 46.564666
			},
			"name": "Carly Compton",
			"notExist": "haha",
			"registered": "2015-09-18T12:59:51Z",
			"tags": ["incididunt", "dolore"]
		},
		"_timestamp": "2018-12-31T22:51:00.118-05:00",
		"_signature": "6b2064ddf73f7b96559ecae424b3b657d1daf62078305e92af991c22e04808d476e8161ec7be58324b662042965a9935de8fb697eb3df7afad5d1885f129f666"
	}, {
		"_id": "5aa8894ed1babae4938aa7ea34b5ee082bf0851e48174815c6d7be0a294afc30",
		"_blockId": "aada5dc20d7de880dc6463dc93033019dfe5a17064b608c3f16d76ef54b8e4dd",
		"_source": {
			"age": 28,
			"friends": [{
				"id": 0,
				"name": "Jimenez Byers"
			}, {
				"id": 1,
				"name": "Gabriela Mayer"
			}],
			"gender": "male",
			"guid": "f51b68c5-f274-4ce1-984f-b4fb4d618ff3",
			"id": "5bf1d3fdf6fd4a5c4638f64e",
			"isActive": false,
			"location": {
				"lat": 53.15213,
				"lon": 46.564666
			},
			"name": "Carly Compton",
			"notExist": "haha",
			"registered": "2015-09-18T12:59:51Z",
			"tags": ["incididunt", "dolore"]
		},
		"_timestamp": "2018-12-31T22:50:33.988-05:00",
		"_signature": "6b2064ddf73f7b96559ecae424b3b657d1daf62078305e92af991c22e04808d476e8161ec7be58324b662042965a9935de8fb697eb3df7afad5d1885f129f666"
	}]
}
```
The response includes the following system fields: `_id` (transaction ID), `_blockId`, `_source`, `_timestamp` and `_signature`.

### Verify a document
__Blocace__ maintains a [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree) for each block. The client is able to use the tree to verify if a document is included in a block and also its integrity in a trustless environment.

```javascript
axios.request({
    url: 'http://localhost:6899/verification/33145e8dfc65b7a6a48f597e13574a32da939942c04f7c5b167b27417e13cb46/82a9c23396dcf047f01cca6923541350a1b90fd37274c63a881a19ba1e97e0da',
    method: 'get',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE2Nzc5LCJpYXQiOjE1NDYzMTYxNzksImlzcyI6ImJsb2Nhc2UifQ.AvARyrxY8bL1M6wzDMNdQ3Yg80nWN92ae2i_x6Rd2LM'}
}).then(response => {
    if (response.data) {
        var verificationPath = response.data.verificationPath
        var txId = "82a9c23396dcf047f01cca6923541350a1b90fd37274c63a881a19ba1e97e0da" // also document id
        var keys = Object.keys(verificationPath)
        var hashData = Buffer.from(txId, "hex")

        for (let i = keys.length - 1; i > 0; i--) {
            var secondHash = Buffer.from(verificationPath[keys[i]], "hex")
        
            if (keys[i]%2 == 0) { // right child
                var prevHashes = Buffer.concat([hashData, secondHash])
                hashData = ethUtil.keccak(prevHashes, 256)
            } else {
                var prevHashes = Buffer.concat([secondHash, hashData])
                hashData = ethUtil.keccak(prevHashes, 256)
            }
        }
        
        console.log("Verified document successfully: " + hashData.equals(Buffer.from(verificationPath["0"], "hex")));
    }
}).catch(error => {
    console.log(error)
})
```

`/verification/{blockID}/{transactionID}` response:
```javascript
{
	"status": "ok",
	"verificationPath": {
		"0": "967d211a165de72e5c8e25e780eec1c573dd44b574cf60a40a62b8253f65de73",
		"1": "99e1953a3eaf7c6779b18bcf6dc939d0b258b9d95280dcc6f5b6fe7a99a5bf2e",
		"6": "768b72fad6299c1b12f1df73f377eb2dbb91abd8dd55d491181c215124269fce",
		"12": "1dae7d2eaffe5596768b318b783e6d40f330e202b285baef61f54e2552a887c5",
		"24": "aea3e287b898d8ed3d7a8d7f1242e3fa7837110d63c41895442b6e44b7961353",
		"48": "51a3cee16a7d5b82371b3305f8dc1bd022b8133b3cd2310b97f28e859ed47f38",
		"95": "3d32bd18441ea9dc17f1a23e8d3264d8a176b22d707c2fc7bd375d0f70822a17",
		"193": "8a1feef1f5afedd8f425009d3c0d35d47825ace89207f5682f0d826debb2ab28",
		"390": "9f8c158640e09d7309bb4a449ac5739a57523ddca06b39090f105cc16d82a8a1"
	}
}
```
The `verificationPath` is a sorted a map with tree index and hash value. The tree index is a level-by-level traversal order.


### Get block information

```javascript
axios.request({
    url: 'http://localhost:6899/block/33145e8dfc65b7a6a48f597e13574a32da939942c04f7c5b167b27417e13cb46',
    method: 'get',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE3NDQ2LCJpYXQiOjE1NDYzMTY4NDYsImlzcyI6ImJsb2Nhc2UifQ.Zc-JZMckCcOS-lJsM8SXHAfiua4iufh57eaJCZwTGDc'}
}).then(response => {
    if (response.data) {
        console.log(
        `Response: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/block/{blockID}` response:
```javascript
{
    "blockId": "33145e8dfc65b7a6a48f597e13574a32da939942c04f7c5b167b27417e13cb46",
    "lastBlockId": "aada5dc20d7de880dc6463dc93033019dfe5a17064b608c3f16d76ef54b8e4dd",
    "blockHeight": 5,
    "totalTransactions": 256
}
```

### Get blockchain information

```javascript
axios.request({
    url: 'http://localhost:6899/info',
    method: 'get',
    timeout: 1000,
    headers: {'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJhZGRyZXNzIjoiMHg3MzBDNWRhOWQ1QTdCMzlBRDNCMmUyMjc0NTI1QjVlYjJBOWZhMjhEIiwiYXVkIjoiYmxvY2FzZSB1c2VyIiwiZXhwIjoxNTQ2MzE3NDQ2LCJpYXQiOjE1NDYzMTY4NDYsImlzcyI6ImJsb2Nhc2UifQ.Zc-JZMckCcOS-lJsM8SXHAfiua4iufh57eaJCZwTGDc'}
}).then(response => {
    if (response.data) {
        console.log(
        `Response: ${JSON.stringify(response.data)}`
        )
    }
}).catch(error => {
    console.log(error)
})
```

`/info` response:
```javascript
{
    "newestBlockId": "33145e8dfc65b7a6a48f597e13574a32da939942c04f7c5b167b27417e13cb46",
    "lastHeight": 5,
    "totalTransactions": 462
}
```

# Usage Reference
## Blocace Javascript API Reference
To develop DAPP talking to Blocace server, we create a handy blocace javascript client at [blocace-client](https://www.npmjs.com/package/blocace-client). Check out [example.js](https://github.com/codingpeasant/blocace/blob/master/client/example.js) for the full usage of the client lib.

## Blocace CLI reference
!> Coming soon
## Blocace web API reference
!> Coming soon
