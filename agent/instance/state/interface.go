// Package state implements stores that store the agent state (the instances that are currently allocated on the node).
package state

// A store stores the all the currently allocated instances
type StateStorer interface {
	Add(*Instance, *Config) error
	Remove(id uint32) error
	Get(id uint32) (*Instance, error)
	GetConfig(id uint32) (*Config, error)
	SetContainerId(id uint32, containerId string) (bool, error)
}
