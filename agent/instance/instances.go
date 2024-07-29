package instance

import (
	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/docker/docker/client"
)

// Instances manages the currently allocated instanes on the node.
type Instances struct {
	docker *client.APIClient
	state  *state.StateStorer
}

func NewInstances(docker *client.APIClient) *Instances {
	return &Instances{docker: docker}
}
