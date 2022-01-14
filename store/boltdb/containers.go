package boltdb

import (
	"encoding/json"
	"time"

	"github.com/geekgonecrazy/uberContainer/models"
	"github.com/geekgonecrazy/uberContainer/store"
	bolt "go.etcd.io/bbolt"
)

func (b *boltStore) GetContainers() ([]models.Container, error) {
	tx, err := b.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	cursor := tx.Bucket(containersBucket).Cursor()

	containers := make([]models.Container, 0)
	for k, data := cursor.First(); k != nil; k, data = cursor.Next() {
		var i models.Container
		if err := json.Unmarshal(data, &i); err != nil {
			return nil, err
		}

		containers = append(containers, i)
	}

	return containers, nil
}

func (b *boltStore) GetContainer(key string) (container models.Container, err error) {
	tx, err := b.Begin(false)
	if err != nil {
		return container, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(containersBucket).Get([]byte(key))
	if bytes == nil {
		return container, store.ErrNotFound
	}

	var i models.Container
	if err := json.Unmarshal(bytes, &i); err != nil {
		return container, err
	}

	return i, nil
}

func (b *boltStore) CreateContainer(container *models.Container) error {
	tx, err := b.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(containersBucket)

	container.CreatedAt = time.Now()
	container.ModifiedAt = time.Now()

	buf, err := json.Marshal(container)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(container.Key), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) UpdateContainer(container *models.Container) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	bucket := tx.Bucket(containersBucket)

	container.ModifiedAt = time.Now()

	buf, err := json.Marshal(container)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(container.Key), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) DeleteContainer(key string) error {
	return s.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(containersBucket).Delete([]byte(key))
	})
}
