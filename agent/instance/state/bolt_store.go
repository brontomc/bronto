package state

import (
	"errors"
	"strconv"
	"time"

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

func (s *BoltStateStore) Close() error {
	return s.db.Close()
}

func (s *BoltStateStore) Add(instance *Instance, config *Config) error {
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

func (s *BoltStateStore) Get(id uint32) (*Instance, error) {
	var i Instance

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

func (s *BoltStateStore) List() ([]uint32, error) {
	ids := []uint32{}

	err := s.db.View(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		c := allInstanceBkt.Cursor()
		for id, _ := c.First(); id != nil; id, _ = c.Next() {
			ids = append(ids, idFromBytes(id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *BoltStateStore) GetConfig(id uint32) (*Config, error) {
	var c Config

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

func (s *BoltStateStore) SetContainerId(id uint32, containerId string) error {
	mapper := func(instance *Instance) {
		instance.ContainerId = containerId
	}
	return s.updateInstance(id, mapper)
}

// SetStatus updates the status of the instance.
// If status is Starting then the instance StartTime should be set to Time.Now()
func (s *BoltStateStore) SetStatus(id uint32, status Status) error {
	mapper := func(instance *Instance) {
		instance.Status = status
		if status == Starting {
			now := time.Now()
			instance.StartTime = &now
		}
	}
	return s.updateInstance(id, mapper)
}

func (s *BoltStateStore) updateInstance(id uint32, mapper func(*Instance)) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		allInstanceBkt := tx.Bucket(instanceBucketName)
		instanceBkt := allInstanceBkt.Bucket(idToBytes(id))
		if instanceBkt == nil {
			return errNotFound
		}

		var i Instance
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
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func idToBytes(id uint32) []byte {
	s := strconv.FormatUint(uint64(id), 10)
	return []byte(s)
}

func idFromBytes(b []byte) uint32 {
	s := string(b)
	id, _ := strconv.ParseUint(s, 10, 32)
	return uint32(id)
}
