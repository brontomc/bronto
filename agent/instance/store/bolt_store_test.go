package store

import (
	"os"
	"testing"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestBoltStore(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "agent.db")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	db, err := bbolt.Open(file.Name(), 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s, err := NewBoltStore(db)
	assert.Nil(t, err, "Error while creating a new BoltStore should be nil")

	i, err := s.Get(1238)
	assert.Nil(t, i, "Instance should not exist in store")
	assert.Nil(t, err, "Error while querying store should be nil")

	expectedInstance := instance.Instance{
		Id:          69,
		Status:      instance.Starting,
		ContainerId: "uwuId",
	}
	expectedConfig := instance.Config{
		DataDirectory: "/wtf",
		ServerJar:     "paper.jar",
		Args:          []string{"-help"},
		ListenPort:    2342,
	}

	err = s.Add(&expectedInstance, &expectedConfig)
	assert.Nil(t, err, "Error while adding an instance should be nil")

	queriedInstance, err := s.Get(expectedInstance.Id)
	assert.Nil(t, err, "Error while querying an instance should be nil")
	assert.Equal(t, &expectedInstance, queriedInstance, "Inserted instance must be the same as queried instance")

	queriedConfig, err := s.GetConfig(expectedInstance.Id)
	assert.Nil(t, err, "Error while querying a config should be nil")
	assert.Equal(t, &expectedConfig, queriedConfig, "Inserted config must be the same as queried instance")

	err = s.Remove(expectedInstance.Id)
	assert.Nil(t, err, "Error while removing an instance should be nil")

	i, err = s.Get(expectedInstance.Id)
	assert.Nil(t, i, "Removed instance should not exist in store")
	assert.Nil(t, err, "Error while querying store should be nil")
}
