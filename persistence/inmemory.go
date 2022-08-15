package persistence

import (
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

// DBConn is the connection for the database used - r DB
var DBConn *badger.DB

func init() {
	DBConn = initDB()
}

func initDB() *badger.DB {
	dbConn, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}

	return dbConn
}
