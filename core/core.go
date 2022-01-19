package core

import (
	"github.com/geekgonecrazy/uberContainer/config"
	"github.com/geekgonecrazy/uberContainer/core/s3"
	"github.com/geekgonecrazy/uberContainer/store"
	"github.com/geekgonecrazy/uberContainer/store/boltdb"
)

var _store store.Store
var _storage *s3.S3Client

var (
	containerDirectory = "./tmp"
)

func Init() {
	//connectionString := "mongodb://localhost:27017/uber"

	/*mongoStore, err := mongo.New(connectionString)
	if err != nil {
		panic(err)
	}

	_store = mongoStore*/

	boltStore, err := boltdb.New("./bolt.db")
	if err != nil {
		panic(err)
	}

	_store = boltStore

	conf, err := config.Get()
	if err != nil {
		panic(err)
	}

	_storage = s3.NewClient(conf.S3)
}
