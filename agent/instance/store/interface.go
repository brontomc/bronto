package store

import (
	"github.com/brontomc/bronto/agent/instance"
)

// A store stores the all the currently allocated instances
type Storer interface {
	Add(*instance.Instance, *instance.Config) error
	Remove(id uint32) error
	Get(id uint32) (*instance.Instance, error)
	GetConfig(id uint32) (*instance.Config, error)
}
