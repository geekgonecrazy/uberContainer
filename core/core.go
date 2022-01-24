package core

import (
	"os"

	"github.com/geekgonecrazy/uberContainer/config"
	"github.com/geekgonecrazy/uberContainer/core/s3"
	"github.com/geekgonecrazy/uberContainer/store"
	"github.com/geekgonecrazy/uberContainer/store/boltdb"
)

var _store store.Store
var _storage *s3.S3Client
var _config *config.Config

var (
	previewTempDirectory = "./tmp"
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

	_config = conf

	if _config.S3.TempFileLocation != "" {
		previewTempDirectory = _config.S3.TempFileLocation
	}

	if _, err := os.Stat(previewTempDirectory); err != nil {
		panic("Please make sure to set valid temp directory")
	}

	_storage = s3.NewClient(conf.S3)
}
