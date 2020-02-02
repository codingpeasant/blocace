<a href="https://www.blocace.com">
	<img width="300" src="./blocace-full-logo.png" alt="blocace Logo" />
</a>
<hr/>

[![Build Status](https://travis-ci.com/codingpeasant/blocace.svg?branch=master)](https://travis-ci.com/codingpeasant/blocace) [![GoDoc](https://godoc.org/github.com/codingpeasant/blocace?status.svg)](https://godoc.org/github.com/codingpeasant/blocace) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

__Blocace__ is a distributed document database powered by the blockchain technology.

## Note to Developers
* This is a prototype.
* The APIs are constantly evolving and designed to demonstrate types of functionality. Expect substantial changes before the release.

## Install

### Compile on Linux/macOS/Windows
> Prerequisite: Go version: 1.12 or later; GCC 5.1 or later.
> 
> Windows may need to install [GCC](http://tdm-gcc.tdragon.net/download) if missing before installing the dependencies. Linux may also need to install gcc using the corresponding package management tool, like `yum install gcc` on RedHat or alike. macOS may need to install [Xcode Command Line Tools](https://www.ics.uci.edu/~pattis/common/handouts/macmingweclipse/allexperimental/macxcodecommandlinetools.html).

Build and run with Go Modules
```bash
git clone https://github.com/codingpeasant/blocace.git
cd blocace
export GO111MODULE=on # Go 1.12 and earlier
go get
go build -ldflags="-s -w -X main.version=0.0.1"
./blocace server
```

You can also use the old way to get dependencies and run
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
go build -ldflags="-s -w -X main.version=0.0.1"
./blocace server
```
### Download for Linux/macOS/Windows
If you'd like to try Blocace directly, please navigate to the [releases](https://github.com/codingpeasant/blocace/releases)

## Docs
Checkout [Blocace In 10 Minutes](https://blocace.com/docs/#/) and [Blocace APIs Reference](https://blocace.com/docs/#/?id=usage-reference)

## License
Blocace is licensed as [Apache 2.0](https://github.com/codingpeasant/blocace/blob/master/LICENSE).
