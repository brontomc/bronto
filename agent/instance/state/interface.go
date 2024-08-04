// Package state implements stores that store the agent state (the instances that are currently allocated on the node).
package state

// A store stores the all the currently allocated instances
type StateStorer interface {
	Add(*Instance, *Config) error
	Remove(id uint32) error
	Get(id uint32) (*Instance, error)
	List() ([]uint32, error) // List all the instance id's
	GetConfig(id uint32) (*Config, error)
	SetContainerId(id uint32, containerId string) error
	// SetStatus updates the status of the instance.
	// If status is Starting then the instance StartTime should be set to Time.Now()
	SetStatus(id uint32, status Status) error
}
