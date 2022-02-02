package core

import (
	"os"

	"github.com/geekgonecrazy/uberContainer/config"
	"github.com/geekgonecrazy/uberContainer/core/s3"
	"github.com/geekgonecrazy/uberContainer/store"
	"github.com/geekgonecrazy/uberContainer/store/boltdb"
	"github.com/geekgonecrazy/uberContainer/store/mongo"
)

var _store store.Store
var _storage *s3.S3Client
var _config *config.Config

var (
	previewTempDirectory = "./tmp"
)

func Init() {
	conf, err := config.Get()
	if err != nil {
		panic(err)
	}

	_config = conf

	switch conf.Database.Type {
	case "mongo":
		if conf.Database.ConnectionString == "" {
			conf.Database.ConnectionString = "mongodb://localhost:27017/uber"
		}

		mongoStore, err := mongo.New(conf.Database.ConnectionString)
		if err != nil {
			panic(err)
		}

		_store = mongoStore
	case "bolt":
		if conf.Database.BoltPath == "" {
			conf.Database.BoltPath = "./bolt.db"
		}

		boltStore, err := boltdb.New(conf.Database.BoltPath)
		if err != nil {
			panic(err)
		}

		_store = boltStore
	default:
		panic("Invalid database type")
	}

	if _config.S3.TempFileLocation != "" {
		previewTempDirectory = _config.S3.TempFileLocation
	}

	if _, err := os.Stat(previewTempDirectory); err != nil {
		panic("Please make sure to set valid temp directory")
	}

	_storage = s3.NewClient(conf.S3)
}
