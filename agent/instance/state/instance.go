package state

import "time"

type Status string

const (
	Offline  Status = "Offline"
	Starting Status = "Starting"
	Running  Status = "Online"
	Error    Status = "Error"
)

type Config struct {
	// DataDirectory represents the root directory of the server.
	DataDirectory string
	// ServerJar represents the file path to the minecraft server executable .jar file.
	ServerJar string
	// Args represents the command-line arguments passed to the JVM.
	Args []string
	// ListenPort represents the port on which the minecraft server should be exposed on the host.
	ListenPort int
}

type Instance struct {
	// Id represents the internal id of the instance.
	Id uint32
	// Status represents the current status of the instance.
	Status Status
	// Time when the server was started or nil if the instance is currently not running.
	StartTime *time.Time
	// ContainerId represents the id of the container in which the server is running.
	// If no container has been created yet ContainerId is empty.
	ContainerId string
}
