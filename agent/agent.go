package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/docker/docker/client"
	"go.etcd.io/bbolt"
)

func Start() error {
	docker, err := client.NewClientWithOpts(client.WithHostFromEnv())
	if err != nil {
		return fmt.Errorf("error while connecting to the docker engine: %w", err)
	}

	bolt, err := bbolt.Open("test.db", 0600, nil)
	if err != nil {
		return fmt.Errorf("error while opening boltdb: %w", err)
	}
	stateStore, err := state.NewBoltStateStore(bolt)
	if err != nil {
		return fmt.Errorf("error while constructing state store: %w", err)
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
		return fmt.Errorf("error while creating instance: %w", err)
	}

	err = instances.StartInstance(0)
	if err != nil {
		return fmt.Errorf("error while starting instance: %w", err)
	}

	logCh, err := instances.Logs(0)
	if err != nil {
		return fmt.Errorf("error while reading logs: %w", err)
	}
	go func() {
		for l := range logCh {
			fmt.Print(l)
		}
	}()
	go func() {
		for {
			r := bufio.NewReader(os.Stdin)
			l, err := r.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				slog.Error("Received error while reading from attached container logs", "error", err)
			}
			instances.SendCommand(0, l)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	fmt.Println("Received Ctrl+C")

	err = instances.StopInstance(0)
	if err != nil {
		return fmt.Errorf("error while stopping instance: %w", err)
	}

	err = instances.RemoveInstance(0)
	if err != nil {
		return fmt.Errorf("error while removing instance: %w", err)
	}
	return nil
}
