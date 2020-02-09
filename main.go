// Blocace is a distributed document database powered by the blockchain technology.
// A super light-weight yet powerful document-oriented database backed by blockchain / distributed ledger technology.
// Data immutable and verifiable is all about trust, which creates the most efficient business.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/codingpeasant/blocace/blockchain"
	"github.com/codingpeasant/blocace/webapi"
)

var secret = "blocace_secret"
var dataDir string
var maxTxsPerBlock int
var maxTimeToGenerateBlock int // milliseconds
var port string
var version string
var loglevel string

func init() {
	fmt.Printf(`
		 ____  __     __    ___   __    ___  ____ 
		(  _ \(  )   /  \  / __) / _\  / __)(  __)
		 ) _ (/ (_/\(  O )( (__ /    \( (__ ) _) 
		(____/\____/ \__/  \___)\_/\_/ \___)(____)

			Community Edition %s

`, version)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "Blocace Community Edition"
	app.Version = version
	app.Copyright = "(c) 2020 Blocace Labs"
	app.Usage = "The Generic Blockchain Solution"
	app.HelpName = "blocace"

	app.Commands = []cli.Command{
		{
			Name:     "server",
			Aliases:  []string{"s"},
			Usage:    "start the major blocace server",
			HelpName: "blocace server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "dir, d",
					Value:       "data",
					Usage:       "the path to the folder of data persistency",
					Destination: &dataDir,
				},
				cli.StringFlag{
					Name:        "secret, s",
					Usage:       "the password to encrypt data and manage JWT",
					Destination: &secret,
				},
				cli.IntFlag{
					Name:        "maxtx, m",
					Value:       2048,
					Usage:       "the max transactions in a block",
					Destination: &maxTxsPerBlock,
				},
				cli.IntFlag{
					Name:        "maxtime, t",
					Value:       2000,
					Usage:       "the time in milliseconds interval to generate a block",
					Destination: &maxTimeToGenerateBlock,
				},
				cli.StringFlag{
					Name:        "port, p",
					Value:       "6899",
					Usage:       "the port that the server listens on",
					Destination: &port,
				},
				cli.StringFlag{
					Name:        "loglevel, l",
					Value:       "info",
					Usage:       "the log levels: panic, fatal, error, warn, info, debug, trace",
					Destination: &loglevel,
				},
			},
			Action: func(c *cli.Context) error {
				switch level := loglevel; level {
				case "panic":
					log.SetLevel(log.PanicLevel)
				case "fatal":
					log.SetLevel(log.PanicLevel)
				case "error":
					log.SetLevel(log.ErrorLevel)
				case "warn":
					log.SetLevel(log.WarnLevel)
				case "info":
					log.SetLevel(log.InfoLevel)
				case "debug":
					log.SetLevel(log.DebugLevel)
				case "trace":
					log.SetLevel(log.TraceLevel)
				default:
					log.SetLevel(log.InfoLevel)
				}

				log.WithFields(log.Fields{
					"path":     dataDir,
					"maxtx":    maxTxsPerBlock,
					"maxtime":  maxTimeToGenerateBlock,
					"port":     port,
					"loglevel": loglevel,
				}).Info("configurations: ")
				server()
				return nil
			},
		},
		{
			Name:     "keygen",
			Aliases:  []string{"k"},
			Usage:    "generate and register an admin account",
			HelpName: "blocace keygen",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "dir, d",
					Value:       "data",
					Usage:       "the path to the folder of data persistency",
					Destination: &dataDir,
				},
			},
			Action: func(c *cli.Context) error {
				keygen()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func server() {
	var bc *blockchain.Blockchain
	var r *blockchain.Receiver
	var dbFile = dataDir + filepath.Dir("/") + "blockchain.db"

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, os.ModePerm)
	}

	if blockchain.DbExists(dbFile) {
		log.Info("db file exists.")
		bc = blockchain.NewBlockchain(dbFile, dataDir)
	} else {
		log.Info("cannot find the db file. creating new...")
		bc = blockchain.CreateBlockchain(dbFile, dataDir)
		generateAdminAccount(bc.Db)
	}

	r = blockchain.NewReceiver(bc, maxTxsPerBlock, maxTimeToGenerateBlock)
	go r.Monitor()

	httpHandler := webapi.NewHTTPHandler(bc, r, secret, version)
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(webapi.ErrorHandler)
	router.Handle("/", httpHandler)
	router.HandleFunc("/jwt", httpHandler.HandleJWT).Methods("POST", "GET")
	router.HandleFunc("/jwt/challenge/{address}", httpHandler.JWTChallenge).Methods("GET")
	router.HandleFunc("/info", httpHandler.HandleInfo).Methods("GET")                                // user
	router.HandleFunc("/block/{blockId}", httpHandler.HandleBlockInfo).Methods("GET")                // user
	router.HandleFunc("/verification/{blockId}/{txId}", httpHandler.HandleMerklePath).Methods("GET") // user
	router.HandleFunc("/search/{collection}", httpHandler.HandleSearch).Methods("POST", "GET")       // user
	router.HandleFunc("/document/{collection}", httpHandler.HandleTransaction).Methods("POST")       // user
	router.HandleFunc("/bulk/{collection}", httpHandler.HandleTransactionBulk).Methods("POST")       // everyone
	router.HandleFunc("/collection", httpHandler.CollectionMappingCreation).Methods("POST")          // admin
	router.HandleFunc("/collections", httpHandler.CollectionList).Methods("GET")                     // user
	router.HandleFunc("/collection/{name}", httpHandler.CollectionMappingGet).Methods("GET")         // user
	router.HandleFunc("/account", httpHandler.AccountRegistration).Methods("POST")
	router.HandleFunc("/account/{address}", httpHandler.AccountUpdate).Methods("POST")                    // admin
	router.HandleFunc("/account/{address}", httpHandler.AccountGet).Methods("GET")                        // user
	router.HandleFunc("/setaccountpermission/{address}", httpHandler.SetAccountReadWrite).Methods("POST") // admin

	handler := cors.Default().Handler(router)

	server := &http.Server{Addr: ":" + port, Handler: handler}

	go func() {
		if err := server.ListenAndServe(); err == nil {
			log.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)

		done <- true
	}()

	log.Info("awaiting signal...")
	<-done

	if err := server.Shutdown(ctx); err != nil {
		log.Error(err)
	}
	log.Info("exiting...")
}

func keygen() {
	if blockchain.DbExists(dataDir + filepath.Dir("/") + "blockchain.db") {
		log.Info("db file exists. generating an admin keypair and registering an account...")
		db, err := bolt.Open(dataDir+filepath.Dir("/")+"blockchain.db", 0600, nil)
		if err != nil {
			log.Panic(err)
		}

		generateAdminAccount(db)

	} else {
		log.Panic("cannot find the db file. please run blocace server first to create the database")
	}
}

func generateAdminAccount(db *bolt.DB) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Panic(err)
	}

	pubKey := privKey.PublicKey
	account := blockchain.Account{Role: blockchain.Role{Name: "admin"}, PublicKey: "04" + fmt.Sprintf("%x", pubKey.X) + fmt.Sprintf("%x", pubKey.Y)}
	addressBytes := []byte(crypto.PubkeyToAddress(pubKey).String())

	result := account.Serialize()

	err = db.Update(func(dbtx *bolt.Tx) error {
		aBucket, _ := dbtx.CreateBucketIfNotExists([]byte(blockchain.AccountsBucket))
		err := aBucket.Put(addressBytes, result)
		if err != nil {
			log.Error(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	log.Info("the account has been created and registered successfully")

	fmt.Printf("\n####################\nPRIVATE KEY: %x\nWARNING: THIS PRIVATE KEY ONLY SHOWS ONCE. PLEASE SAVE IT NOW AND KEEP IT SAFE. YOU ARE THE ONLY PERSON THAT IS SUPPOSED TO OWN THIS KEY IN THE WORLD.\n####################\n\n", privKey.D)

	// priv, err := crypto.HexToECDSA(fmt.Sprintf("%x", privKey.D))
	// if err != nil {
	// 	log.Error(err)
	// }
	// fmt.Printf("%x\n%x\n", priv.PublicKey.X, priv.PublicKey.Y)
}
