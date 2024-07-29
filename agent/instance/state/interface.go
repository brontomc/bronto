// Package state implements stores that store the agent state (the instances that are currently allocated on the node)
package state

import (
	"github.com/brontomc/bronto/agent/instance"
)

// A store stores the all the currently allocated instances
type StateStorer interface {
	Add(*instance.Instance, *instance.Config) error
	Remove(id uint32) error
	Get(id uint32) (*instance.Instance, error)
	GetConfig(id uint32) (*instance.Config, error)
	SetContainerId(id uint32, containerId string) (bool, error)
}
