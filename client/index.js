const ethUtil = require('ethereumjs-util')
const ethWallet = require('ethereumjs-wallet')
const axios = require('axios')
const crypto = require('crypto')
const secp256k1 = require('secp256k1')

const httpRequestTimeout = 5000 // ms

class Blocase {
  constructor (wallet, hostname, port, protocol) {
    this.wallet = wallet
    this.hostname = hostname || 'localhost'
    this.port = port || 6899
    this.protocol = protocol || 'http'
  }

  static create (hostname, port, protocol) {
    return new this(ethWallet.generate(), hostname, port, protocol)
  }

  static createFromPrivateKey (privKey, hostname, port, protocol) {
    if (!ethUtil.isValidPrivate(Buffer.from(privKey, 'hex'))) {
      throw new Error('invalid private key')
    }
    return new this(ethWallet.fromPrivateKey(Buffer.from(privKey, 'hex')), hostname, port, protocol)
  }

  encryptPrivateKey (password) {
    try {
      const cipher = crypto.createCipher('aes-256-cbc', password)
      const encrypted = Buffer.concat([cipher.update(Buffer.from(JSON.stringify(this.wallet.getPrivateKey().toString('hex')), 'utf8')), cipher.final()])
      return encrypted
    } catch (exception) {
      throw new Error(exception.message)
    }
  }

  static decryptPrivateKey (encrypted, password) {
    try {
      const decipher = crypto.createDecipher('aes-256-cbc', password)
      const decrypted = Buffer.concat([decipher.update(encrypted), decipher.final()])
      return JSON.parse(decrypted.toString())
    } catch (exception) {
      throw new Error(exception.message)
    }
  }

  static verifySignature (rawDocument, signature, publicKey) {
    const docHash = ethUtil.keccak(rawDocument, 256)
    return secp256k1.verify(docHash, Buffer.from(signature, 'hex'), Buffer.from('04' + publicKey, 'hex'))
  }

  async createAccount (accountPayload) {
    // accountPayload.publicKey = this.wallet.getPublicKey().toString('hex')
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account',
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: accountPayload
    })

    return accountRes
  }

  async updateAccount (accountPayload, address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account/' + address,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: accountPayload
    })

    return accountRes
  }

  async setAccountReadWrite (permissionPayload, address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/setaccountpermission/' + address,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: permissionPayload
    })

    return accountRes
  }

  async getChallenge () {
    return axios.get(this.protocol + '://' + this.hostname + ':' + this.port + '/jwt/challenge/' + this.wallet.getChecksumAddressString())
  }

  async getJWT () {
    const challengeResponse = await this.getChallenge()
    var challengeHash = ethUtil.keccak(challengeResponse.data.challenge, 256)
    var sig = ethUtil.ecsign(challengeHash, Buffer.from(this.wallet.getPrivateKey(), 'hex'))

    const jwt = await axios.post(this.protocol + '://' + this.hostname + ':' + this.port + '/jwt', {
      'address': this.wallet.getChecksumAddressString(),
      'signature': sig.r.toString('hex') + sig.s.toString('hex')
    })

    this.token = jwt.data.token
    return jwt.data.token
  }

  async getAccount (address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account/' + address,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    return accountRes.data
  }

  async createCollection (collectionPayload) {
    const collectionCreationRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collection',
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: collectionPayload
    })

    return collectionCreationRes.data
  }

  async signAndPutDocument (document, collection) {
    const docHash = ethUtil.keccak(JSON.stringify(document), 256)
    const sig = ethUtil.ecsign(docHash, Buffer.from(this.wallet.getPrivateKey(), 'hex'))

    const putDocRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/document/' + collection,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: {
        'rawDocument': JSON.stringify(document),
        'signature': sig.r.toString('hex') + sig.s.toString('hex')
      }
    })

    return putDocRes.data
  }

  async query (queryPayload, collection) {
    const queryRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/search/' + collection,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token},
      data: queryPayload
    })

    return queryRes.data
  }

  async verifyTransaction (blockId, transationId) {
    const verificationPathRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/verification/' + blockId + '/' + transationId,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    const verificationPath = verificationPathRes.data.verificationPath
    var keys = Object.keys(verificationPath)
    var hashData = Buffer.from(transationId, 'hex')

    for (let i = keys.length - 1; i > 0; i--) {
      var secondHash = Buffer.from(verificationPath[keys[i]], 'hex')
      if (keys[i] % 2 === 0) { // right child
        var prevHashes = Buffer.concat([hashData, secondHash])
        hashData = ethUtil.keccak(prevHashes, 256)
      } else {
        prevHashes = Buffer.concat([secondHash, hashData])
        hashData = ethUtil.keccak(prevHashes, 256)
      }
    }

    return hashData.equals(Buffer.from(verificationPath['0'], 'hex'))
  }

  async getBlockInfo (blockId) {
    const blockInfoRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/block/' + blockId,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    return blockInfoRes.data
  }

  async getBlockchainInfo () {
    const blockchainRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/info',
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    return blockchainRes.data
  }

  async getCollections () {
    const collectionsRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collections',
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    return collectionsRes.data
  }

  async getCollection (collectionName) {
    const collectionRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collection/' + collectionName,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: {'Authorization': 'Bearer ' + this.token}
    })

    return collectionRes.data
  }
}

module.exports = Blocase
