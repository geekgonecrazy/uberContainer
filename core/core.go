package core

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/FideTechSolutions/uberContainer/config"
	"github.com/FideTechSolutions/uberContainer/core/s3"
	"github.com/FideTechSolutions/uberContainer/store"
	"github.com/FideTechSolutions/uberContainer/store/boltdb"
	"github.com/FideTechSolutions/uberContainer/store/mongo"
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

	// If not added the image detection will always show unknown format
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

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
		if os.IsNotExist(err) {
			errDir := os.MkdirAll(previewTempDirectory, 0755)
			if errDir != nil {
				panic(err)
			}
		}
	}

	_storage = s3.NewClient(conf.S3)
}
