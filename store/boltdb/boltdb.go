package boltdb

import (
	"time"

	"github.com/geekgonecrazy/uberContainer/store"
	bolt "go.etcd.io/bbolt"
)

type boltStore struct {
	*bolt.DB
}

var (
	containersBucket = []byte("containers")
)

func New(file string) (store.Store, error) {
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 15 * time.Second})
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists(containersBucket); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &boltStore{db}, nil
}

func (b *boltStore) CheckDb() error {
	tx, err := b.Begin(false)
	if err != nil {
		return err
	}

	return tx.Rollback()
}
