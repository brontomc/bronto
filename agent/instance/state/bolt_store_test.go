package state

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestBoltStore(t *testing.T) {
	file, err := os.CreateTemp("", "agent.db")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer os.Remove(file.Name())

	db, err := bbolt.Open(file.Name(), 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s, err := NewBoltStateStore(db)
	assert.Nil(t, err, "Error while creating a new BoltStore should be nil")

	i, err := s.Get(1238)
	assert.Nil(t, i, "Instance should not exist in store")
	assert.Nil(t, err, "Error while querying store should be nil")

	expectedInstance := Instance{
		Id:          69,
		Status:      Starting,
		ContainerId: "uwuId",
	}
	expectedConfig := Config{
		DataDirectory: "/wtf",
		ServerJar:     "paper.jar",
		Args:          []string{"-help"},
		ListenPort:    2342,
	}

	err = s.Add(&expectedInstance, &expectedConfig)
	assert.Nil(t, err, "Error while adding an instance should be nil")

	queriedInstance, err := s.Get(expectedInstance.Id)
	assert.Equal(t, &expectedInstance, queriedInstance, "Inserted instance must be the same as queried instance")
	assert.Nil(t, err, "Error while querying an instance should be nil")

	queriedConfig, err := s.GetConfig(expectedInstance.Id)
	assert.Equal(t, &expectedConfig, queriedConfig, "Inserted config must be the same as queried instance")
	assert.Nil(t, err, "Error while querying a config should be nil")

	newContainerId := "UwuUpdated"
	err = s.SetContainerId(expectedInstance.Id, newContainerId)
	assert.Nil(t, err, "Error while setting the container id should be nil")
	queriedInstance, err = s.Get(expectedInstance.Id)
	assert.Equal(t, newContainerId, queriedInstance.ContainerId, "New container id should be the same as the queried one")
	assert.Nil(t, err, "Error while querying an instance should be nil")

	err = s.Remove(expectedInstance.Id)
	assert.Nil(t, err, "Error while removing an instance should be nil")

	i, err = s.Get(expectedInstance.Id)
	assert.Nil(t, i, "Removed instance should not exist in store")
	assert.Nil(t, err, "Error while querying store should be nil")
}
