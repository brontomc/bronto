package state

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

// A BoltStateStore uses boltdb to implement the Storer interface.
// All entities are stored using the msgpack encoding.
type BoltStateStore struct {
	db *bolt.DB
}

func NewBoltStateStore(db *bolt.DB) (*BoltStateStore, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(instanceBucketName)
		return err
	})
	return &BoltStateStore{db: db}, err
}

func (s *BoltStateStore) Add(instance *instance.Instance, config *instance.Config) error {
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

func (s *BoltStateStore) Remove(id uint32) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		allInstanceBkt.DeleteBucket(idToBytes(id))
		return nil
	})

	return err
}

func (s *BoltStateStore) Get(id uint32) (*instance.Instance, error) {
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

func (s *BoltStateStore) GetConfig(id uint32) (*instance.Config, error) {
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

func (s *BoltStateStore) SetContainerId(id uint32, containerId string) (bool, error) {
	mapper := func(instance *instance.Instance) {
		instance.ContainerId = containerId
	}
	return s.updateInstance(id, mapper)
}

func (s *BoltStateStore) updateInstance(id uint32, mapper func(*instance.Instance)) (bool, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		instanceBkt := allInstanceBkt.Bucket(idToBytes(id))
		if instanceBkt == nil {
			return errNotFound
		}

		var i instance.Instance
		data := instanceBkt.Get(instanceKey)
		err := msgpack.Unmarshal(data, &i)
		if err != nil {
			return err
		}

		mapper(&i)

		data, err = msgpack.Marshal(&i)
		if err != nil {
			return err
		}
		err = instanceBkt.Put(instanceKey, data)
		if err != nil {
			return err
		}

		return nil
	})

	if errors.Is(err, errNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func idToBytes(id uint32) []byte {
	s := strconv.FormatUint(uint64(id), 10)
	return []byte(s)
}
