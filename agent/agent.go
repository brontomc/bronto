package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/docker/docker/client"
	"go.etcd.io/bbolt"
)

func Start() {
	docker, err := client.NewClientWithOpts(client.WithHostFromEnv())
	if err != nil {
		log.Fatalf("Error while connecting to the docker engine: %s", err)
	}

	bolt, err := bbolt.Open("test.db", 0600, nil)
	if err != nil {
		log.Fatalf("Error while opening boltdb: %s", err)
	}
	stateStore, err := state.NewBoltStateStore(bolt)
	if err != nil {
		log.Fatalf("Error while constructing state store: %s", err)
	}
	defer stateStore.Close()
	defer os.Remove("test.db")

	instances := instance.NewInstances(context.Background(), docker, stateStore)

	instanceConfig := state.Config{
		DataDirectory: "testServer",
		ServerJar:     "paper.jar",
		Args:          []string{"-nogui"},
		ListenPort:    25565,
	}
	err = instances.AddInstance(0, &instanceConfig)
	if err != nil {
		log.Fatalf("Error while creating instance: %s", err)
	}

	err = instances.StartInstance(0)
	if err != nil {
		log.Fatalf("Error while starting instance: %s", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	fmt.Println("Received Ctrl+C")

	err = instances.StopInstance(0)
	if err != nil {
		log.Fatalf("Error while stopping instance: %s", err)
	}

	err = instances.RemoveInstance(0)
	if err != nil {
		log.Fatalf("Error while removing instance: %s", err)
	}
}
