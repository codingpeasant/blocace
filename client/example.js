const Blocace = require('./index.js')

// initializing the default admin account
// Replace the private key with the Blocase server admin key:
// ####################
// PRIVATE KEY: 39b7c93247a6fd26e7f6f357030c76fc20263a55e9f2e5cd5e80f1285233c936
// WARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.
// ####################
var blocace = Blocace.createFromPrivateKey('39b7c93247a6fd26e7f6f357030c76fc20263a55e9f2e5cd5e80f1285233c936')

// encrypt and decrypt the seed
var encryptPrivKey = blocace.encryptPrivateKey('123456')
var decryptPrivKey = Blocace.decryptPrivateKey(encryptPrivKey, '123456')

console.log('Private key decrypted: ' + (blocace.wallet.privateKey === decryptPrivKey) + '\n')

// user account payload
const accountPayload = {
  'dateOfBirth': '2018-10-01',
  'firstName': 'Hooper',
  'lastName': 'Vincent',
  'company': 'MITROC',
  'position': 'VP of Marketing',
  'email': 'hoopervincent@mitroc.com',
  'phone': '+1 (849) 503-2756',
  'address': '699 Canton Court, Mulino, South Dakota, 9647',
  'publicKey': 'b0a303c71d99ad217c77af1e4d5b85e3ccc3e359d2ac9ff95e042fb0e0016e4d4c25482ba57de472c44c58f6fb124a0ab86613b0dcd1253a23d5ae00180854fa'
}

const collectionMappingPaylod = {
  'collection': 'new1',
  'fields': {
    'id': { 'type': 'text' },
    'guid': { 'type': 'text' },
    'age': { 'type': 'number', 'encrypted': true },
    'registered': { 'type': 'datetime' },
    'isActive': { 'type': 'boolean' },
    'gender': { 'type': 'text' },
    'name': { 'type': 'text', 'encrypted': true },
    'location': { 'type': 'geopoint' },
    'tags': { 'type': 'text' }
  }
}

const permission = {
  "collectionsWrite": ["default", "new1"],
  "collectionsReadOverride": ["default", "collection2"]
}

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

const queryPayload = {
  'size': 10,
  'from': 0,
  'query': {
    'match': 'Compton',
    'field': 'name'
  }
}

function timeout(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

async function start() {
  try {
  // get JWT
  const jwt = await blocace.getJWT()
  console.log('JWT (admin): ' + jwt + '\n')

  // register a new user account
  const accountRes = await blocace.createAccount(accountPayload)
  console.log('Address of new account: ' + accountRes.data.address + '\n')

  // get account
  const account = await blocace.getAccount(accountRes.data.address)
  console.log('New account Info: ' + JSON.stringify(account) + '\n')

  // set the new account read / write permission
  const accountPermissionRes = await blocace.setAccountReadWrite(permission, accountRes.data.address)
  console.log('Account permission response: ' + JSON.stringify(accountPermissionRes.data.message) + '\n')

  // get the user account again
  const accountUpdated = await blocace.getAccount(accountRes.data.address)
  console.log('Get the update account: ' + JSON.stringify(accountUpdated))

  // create collection
  const collectionCreationRes = await blocace.createCollection(collectionMappingPaylod)
  console.log('New collection info: ' + JSON.stringify(collectionCreationRes) + '\n')

  // initialize the new user account
  var blocaceUser = Blocace.createFromPrivateKey('277d271593d205c6078964c31fb393303efd76d5297906f60d2a7a7d7d12c99a')
  // get JWT for the user account
  const jwtUser = await blocaceUser.getJWT()
  console.log('JWT (new user): ' + jwtUser + '\n')

  // update account
  accountPayload.email = 'asd@asd.com'
  const accountUpdateRes = await blocaceUser.updateAccount(accountPayload, accountRes.data.address)
  console.log('Update account response: ' + JSON.stringify(accountUpdateRes.data.message) + '\n')

  // put a document
  const putDocRes = await blocaceUser.signAndPutDocument(document, 'new1')
  console.log('Put document response: ' + JSON.stringify(putDocRes) + '\n')

  // wait for block to be generated
  await timeout(2000)
  console.log('Waiting for the document to be included in a block... \n')

  // query the blockchain
  const queryRes = await blocaceUser.query(queryPayload, 'new1')
  console.log('Query result: ' + JSON.stringify(queryRes) + '\n')

  // verify if the transaction is included in the block (by block merkle tree rebuild)
  const verificationPassed = await blocaceUser.verifyTransaction(queryRes.hits[0]._blockId, queryRes.hits[0]._id)
  console.log('Document included in the block: ' + verificationPassed + '\n')

  // verify signature
  console.log('The document\'s integrity check passed: ' + Blocace.verifySignature(queryRes.hits[0]._source, queryRes.hits[0]._signature, blocaceUser.wallet.address) + '\n')

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
  } catch(err) {
    console.error(err.response.data);
  }
}

start()
