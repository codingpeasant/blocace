const ethers = require('ethers');
const axios = require('axios')
const crypto = require('crypto')

const httpRequestTimeout = 5000 // ms
const iv = '5ff3a79b8206d1ebea614790dffbc0ef'; // fixed initial vector to start with

class Blocace {
  constructor(wallet, hostname, port, protocol) {
    this.wallet = wallet
    this.hostname = hostname || 'localhost'
    this.port = port || 6899
    this.protocol = protocol || 'http'
  }

  static create(hostname, port, protocol) {
    return new this(new ethers.Wallet.createRandom(), hostname, port, protocol)
  }

  static createFromPrivateKey(privKey, hostname, port, protocol) {
    return new this(new ethers.Wallet(privKey), hostname, port, protocol)
  }

  encryptPrivateKey(password) {
    try {
      const key = crypto.createHash('sha256').update(String(password)).digest('base64').substr(0, 32);
      const cipher = crypto.createCipheriv('aes-256-cbc', key, Buffer.from(iv, 'hex'))
      const encrypted = Buffer.concat([cipher.update(this.wallet.privateKey), cipher.final()])
      return encrypted
    } catch (exception) {
      throw new Error(exception.message)
    }
  }

  static decryptPrivateKey(encrypted, password) {
    try {
      const key = crypto.createHash('sha256').update(String(password)).digest('base64').substr(0, 32);
      const decipher = crypto.createDecipheriv('aes-256-cbc', key, Buffer.from(iv, 'hex'))
      const decrypted = Buffer.concat([decipher.update(encrypted), decipher.final()])
      return decrypted.toString()
    } catch (exception) {
      throw new Error(exception.message)
    }
  }

  static verifySignature(rawDocument, signature, address) {
    return ethers.utils.recoverAddress(ethers.utils.keccak256(Buffer.from(rawDocument)), '0x' + signature + '00') == address
  }

  async createAccount(accountPayload) {
    // accountPayload.publicKey = this.wallet.getPublicKey().toString('hex')
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account',
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: accountPayload
    })

    return accountRes
  }

  async updateAccount(accountPayload, address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account/' + address,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: accountPayload
    })

    return accountRes
  }

  async setAccountReadWrite(permissionPayload, address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/setaccountpermission/' + address,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: permissionPayload
    })

    return accountRes
  }

  async getChallenge() {
    return axios.get(this.protocol + '://' + this.hostname + ':' + this.port + '/jwt/challenge/' + this.wallet.address)
  }

  async getJWT() {
    const challengeResponse = await this.getChallenge()

    const challengeHash = ethers.utils.keccak256(Buffer.from(challengeResponse.data.challenge))
    const sig = this.wallet.signingKey.signDigest(challengeHash);

    const jwt = await axios.post(this.protocol + '://' + this.hostname + ':' + this.port + '/jwt', {
      'address': this.wallet.address,
      'signature': sig.r.substring(2) + sig.s.substring(2)
    })

    this.token = jwt.data.token
    return jwt.data.token
  }

  async getAccount(address) {
    const accountRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/account/' + address,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    return accountRes.data
  }

  async createCollection(collectionPayload) {
    const collectionCreationRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collection',
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: collectionPayload
    })

    return collectionCreationRes.data
  }

  async signAndPutDocument(document, collection) {
    const docHash = ethers.utils.keccak256(Buffer.from(JSON.stringify(document)))
    const sig = this.wallet.signingKey.signDigest(docHash)

    const putDocRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/document/' + collection,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: {
        'rawDocument': JSON.stringify(document),
        'signature': sig.r.substring(2) + sig.s.substring(2)
      }
    })

    return putDocRes.data
  }

  // WARNING: this makes the document unverifiable
  async putDocumentBulk(documents, collection) {
    const putDocRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/bulk/' + collection,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: documents
    })

    return putDocRes.data
  }

  async query(queryPayload, collection) {
    const queryRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/search/' + collection,
      method: 'post',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token },
      data: queryPayload
    })

    return queryRes.data
  }

  async verifyTransaction(blockId, transationId) {
    const verificationPathRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/verification/' + blockId + '/' + transationId,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    const verificationPath = verificationPathRes.data.verificationPath
    var keys = Object.keys(verificationPath)
    var hashData = Buffer.from(transationId, 'hex')

    for (let i = keys.length - 1; i > 0; i--) {
      var secondHash = Buffer.from(verificationPath[keys[i]], 'hex')
      if (keys[i] % 2 === 0) { // right child
        var prevHashes = Buffer.concat([hashData, secondHash])
        hashData = Buffer.from(ethers.utils.keccak256(prevHashes).substring(2), 'hex')
      } else {
        prevHashes = Buffer.concat([secondHash, hashData])
        hashData = Buffer.from(ethers.utils.keccak256(prevHashes).substring(2), 'hex')
      }
    }

    return hashData.equals(Buffer.from(verificationPath['0'], 'hex'))
  }

  async getBlockInfo(blockId) {
    const blockInfoRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/block/' + blockId,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    return blockInfoRes.data
  }

  async getBlockchainInfo() {
    const blockchainRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/info',
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    return blockchainRes.data
  }

  async getCollections() {
    const collectionsRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collections',
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    return collectionsRes.data
  }

  async getCollection(collectionName) {
    const collectionRes = await axios.request({
      url: this.protocol + '://' + this.hostname + ':' + this.port + '/collection/' + collectionName,
      method: 'get',
      timeout: httpRequestTimeout,
      headers: { 'Authorization': 'Bearer ' + this.token }
    })

    return collectionRes.data
  }
}

module.exports = Blocace
