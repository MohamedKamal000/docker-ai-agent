package rag

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"

	models "docker-cli/internal/rag/models"
)

type BoltClient struct {
	db *bolt.DB
}

const (
	DB_NAME     string = "dockerChunksDb.db"
	BUCKET_NAME string = "ChunksData"
)

func toBytes(data string) []byte {
	return []byte(data)
}

func NewBoltClient(dbPath string) (*BoltClient, error) {
	db, err := bolt.Open(filepath.Join(dbPath, DB_NAME), 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(toBytes(BUCKET_NAME))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &BoltClient{
		db: db,
	}, nil
}

func (cb *BoltClient) ConcurrentInsert(document models.VectorDocument) error {
	return cb.db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(toBytes(BUCKET_NAME))
		data, err := json.Marshal(document)
		if err != nil {
			return err
		}

		return bucket.Put(toBytes(document.Id), data)
	})
}

func (cb *BoltClient) ReadAllDocuments(ch chan<- models.VectorDocument) error {
	defer close(ch)
	return cb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(toBytes(BUCKET_NAME))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var nextValue models.VectorDocument
			err := json.Unmarshal(v, &nextValue)
			if err != nil {
				return err
			}
			ch <- nextValue
		}
		return nil
	})
}

func (cb *BoltClient) ReadDocument(Id string) (models.VectorDocument, error) {
	var result models.VectorDocument
	err := cb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(toBytes(BUCKET_NAME))
		value := bucket.Get(toBytes(Id))
		if value == nil {
			return fmt.Errorf("document with Id %s not found", Id)
		}
		return json.Unmarshal(value, &result)
	})
	if err != nil {
		return models.VectorDocument{}, err
	}
	return result, nil
}

func (cb *BoltClient) Close() {
	if cb.db != nil {
		cb.db.Close()
	}
}
