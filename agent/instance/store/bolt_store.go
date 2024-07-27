package store

import (
	"errors"
	"strconv"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
)

var (
	instanceBucketName = []byte("instances")
	instanceKey        = []byte("instance")
	configKey          = []byte("config")
)

var errNotFound = errors.New("")

// A BoltStore uses boltdb to implement the Storer interface.
// All entities are stored using the msgpack encoding.
type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(db *bolt.DB) (*BoltStore, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(instanceBucketName)
		return err
	})
	return &BoltStore{db: db}, err
}

func (s *BoltStore) Add(instance *instance.Instance, config *instance.Config) error {
	idata, err := msgpack.Marshal(instance)
	if err != nil {
		return err
	}
	cdata, err := msgpack.Marshal(config)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		instanceBkt, err := allInstanceBkt.CreateBucketIfNotExists(idToBytes(instance.Id))
		if err != nil {
			return err
		}

		err = instanceBkt.Put(instanceKey, idata)
		if err != nil {
			return err
		}

		err = instanceBkt.Put(configKey, cdata)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *BoltStore) Remove(id uint32) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		allInstanceBkt.DeleteBucket(idToBytes(id))
		return nil
	})

	return err
}

func (s *BoltStore) Get(id uint32) (*instance.Instance, error) {
	var i instance.Instance

	err := s.db.View(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		instanceBkt := allInstanceBkt.Bucket(idToBytes(id))
		if instanceBkt == nil {
			return errNotFound
		}

		data := instanceBkt.Get(instanceKey)
		return msgpack.Unmarshal(data, &i)
	})

	if errors.Is(err, errNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *BoltStore) GetConfig(id uint32) (*instance.Config, error) {
	var c instance.Config

	err := s.db.View(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		instanceBkt := allInstanceBkt.Bucket(idToBytes(id))
		if instanceBkt == nil {
			return errNotFound
		}

		data := instanceBkt.Get(configKey)
		return msgpack.Unmarshal(data, &c)
	})

	if errors.Is(err, errNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func idToBytes(id uint32) []byte {
	s := strconv.FormatUint(uint64(id), 10)
	return []byte(s)
}
