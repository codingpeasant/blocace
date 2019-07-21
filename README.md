<a href="https://www.blocase.com">
	<img width="300" src="./blocase-full-logo.png" alt="Blocase Logo" />
</a>
<hr/>

__Blocase__ is a distributed NoSQL document database powered by the blockchain technology.

## Note to Developers
* This is a prototype.
* The APIs are constantly evolving and designed to demonstrate types of functionality. Expect substantial changes before the release.

## Install

### Compile on Linux/MacOS
> Prerequisite: Go version: 1.12 or later
```bash
git clone https://github.com/codingpeasant/blocase.git
mkdir -p ~/go/src/github.com/blocase
mv ./blocase ~/go/src/github.com/blocase/
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
cd ~/go/src/github.com/blocase/blocase
go build -ldflags="-s -w"
```
You can also use Go Modules to get the dependencies with `export GO111MODULE=on` and `go get`.

#### OR
### Download

Blocase supports Darwin, Linux and Windows with amd64.

__MacOS__
```bash
wget https://s3.us-east-2.amazonaws.com/blocase/darwin/blocase-v0.0.1-darwin
```
__Linux__
```bash
wget https://s3.us-east-2.amazonaws.com/blocase/linux/blocase-v0.0.1-linux
```
__Windows__
```bash
wget https://s3.us-east-2.amazonaws.com/blocase/windows/blocase-v0.0.1-win64.exe
```

## Docs
Checkout [Blocase In 10 Minutes](https://www.blocase.com/docs/#/)

## License
Blocase is licensed as [Apache 2.0](https://github.com/codingpeasant/blocase/blob/master/LICENSE).