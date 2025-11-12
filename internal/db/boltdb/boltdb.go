package boltdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(path string) (*BoltStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

func (b *BoltStore) Put(bucket string, key string, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		return bucket.Put([]byte(key), value)
	})
}

func (b *BoltStore) Get(bucket string, key string) ([]byte, error) {
	var val []byte

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		v := bucket.Get([]byte(key))
		if v == nil {
			return errors.New("key not found")
		}

		val = make([]byte, len(v))
		copy(val, v)

		return nil
	})

	return val, err
}

func (b *BoltStore) Delete(bucket string, key string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		return bucket.Delete([]byte(key))
	})
}

func (b *BoltStore) ListKeys(bucket string) ([]string, error) {
	var keys []string

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		return bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})

	return keys, err
}

func (b *BoltStore) CreateBucket(bucket string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket %q: %w", bucket, err)
		}
		return nil
	})
}

func (b *BoltStore) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
