<a href="https://www.blocace.com">
	<img width="300" src="./blocace-full-logo.png" alt="blocace Logo" />
</a>
<hr/>

__Blocace__ is a distributed NoSQL document database powered by the blockchain technology.

## Note to Developers
* This is a prototype.
* The APIs are constantly evolving and designed to demonstrate types of functionality. Expect substantial changes before the release.

## Install

### Compile on Linux/MacOS
> Prerequisite: Go version: 1.12 or later
```bash
git clone https://github.com/codingpeasant/blocace.git
mkdir -p ~/go/src/github.com/blocace
mv ./blocace ~/go/src/github.com/blocace/
go get github.com/sirupsen/logrus
go get github.com/dgrijalva/jwt-go
go get github.com/boltdb/bolt
go get github.com/blevesearch/bleve
go get github.com/ethereum/go-ethereum/crypto
go get github.com/gorilla/mux
go get gopkg.in/validator.v2
go get github.com/syndtr/goleveldb/leveldb
go get github.com/urfave/cli
go get github.com/thoas/go-funk
cd ~/go/src/github.com/blocace/blocace
go build -ldflags="-s -w"
```
You can also use Go Modules to get the dependencies with `export GO111MODULE=on` (Go 1.12 and earlier) and `go get`.

## Docs
Checkout [Blocace In 10 Minutes](https://github.com/codingpeasant/blocace/tree/master/docs)

## License
Blocace is licensed as [Apache 2.0](https://github.com/codingpeasant/blocace/blob/master/LICENSE).