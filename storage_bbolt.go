package appserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type BBoltStore struct {
	db *bolt.DB

	bucket []byte
}

func NewBBoltStore(path string) (*BBoltStore, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	b := BBoltStore{
		db:     db,
		bucket: []byte("credentials"),
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(b.bucket)
		if err != nil && !errors.Is(err, bolt.ErrBucketExists) {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *BBoltStore) Store(ctx context.Context, credentials Credentials) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		cred, err := json.Marshal(credentials)
		if err != nil {
			return err
		}

		b := tx.Bucket(s.bucket)
		return b.Put([]byte(credentials.ShopID), cred)
	})
}

func (s *BBoltStore) Get(ctx context.Context, shopID string) (Credentials, error) {
	var credentials Credentials

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		v := b.Get([]byte(shopID))

		return json.Unmarshal(v, &credentials)
	})
	if err != nil {
		return Credentials{}, err
	}

	return credentials, nil
}

func (s *BBoltStore) Delete(ctx context.Context, shopID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		return b.Delete([]byte(shopID))
	})
}

func (s *BBoltStore) Close() {
	s.db.Close()
}
