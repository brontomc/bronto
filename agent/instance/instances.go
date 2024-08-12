package instance

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	ErrInstanceAlreadyExists = errors.New("instance with this id already exists")
	ErrInstanceDoesNotExist  = errors.New("instance with this id does not exist")
	ErrInstanceIsRunning     = errors.New("instance is running")
	ErrInstanceIsNotRunning  = errors.New("instance is not running")
)

// Instances manages the currently allocated instanes on the node.
type Instances struct {
	docker client.APIClient
	state  state.StateStorer

	ctx context.Context
}

func NewInstances(ctx context.Context, docker client.APIClient, stateStore state.StateStorer) *Instances {
	return &Instances{docker: docker, state: stateStore, ctx: ctx}
}

func (i *Instances) AddInstance(id uint32, config *state.Config) error {
	instance, err := i.state.Get(id)
	if err != nil {
		return err
	}
	if instance != nil {
		return ErrInstanceAlreadyExists
	}

	instance = &state.Instance{
		Id:          id,
		Status:      state.Offline,
		ContainerId: "",
	}
	i.state.Add(instance, config)

	return nil
}

func (i *Instances) InstanceStatus(id uint32) (*state.Status, error) {
	instance, err := i.state.Get(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrInstanceDoesNotExist
	}

	return &instance.Status, nil
}

func (i *Instances) GetInstance(id uint32) (*state.Instance, error) {
	instance, err := i.state.Get(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrInstanceDoesNotExist
	}

	return instance, nil
}

func (i *Instances) RemoveInstance(id uint32) error {
	instance, err := i.state.Get(id)
	if err != nil {
		return err
	}
	if instance == nil {
		return nil
	}

	if instance.Status != state.Offline {
		return ErrInstanceIsRunning
	}

	if instance.ContainerId != "" {
		err = i.docker.ContainerRemove(i.ctx, instance.ContainerId, container.RemoveOptions{})
		if err != nil {
			return err
		}
	}

	return i.state.Remove(id)
}

func (i *Instances) StartInstance(id uint32) error {
	instance, err := i.state.Get(id)
	if err != nil {
		return err
	}
	if instance == nil {
		return nil
	}

	if instance.Status != state.Offline {
		return ErrInstanceIsRunning
	}

	cid, err := i.ensureContainer(id)
	if err != nil {
		return err
	}

	err = i.docker.ContainerStart(i.ctx, cid, container.StartOptions{})
	if err != nil {
		return err
	}

	err = i.state.SetStatus(id, state.Running)

	return err
}

func (i *Instances) StopInstance(id uint32) error {
	instance, err := i.state.Get(id)
	if err != nil {
		return err
	}
	if instance == nil {
		return ErrInstanceDoesNotExist
	}

	if instance.Status != state.Running {
		return ErrInstanceIsNotRunning
	}

	timeout := 60
	err = i.docker.ContainerStop(i.ctx, instance.ContainerId, container.StopOptions{Timeout: &timeout})
	if err != nil {
		return err
	}

	err = i.state.SetStatus(id, state.Offline)

	return err
}

func (i *Instances) Logs(id uint32) (<-chan string, error) {
	instance, err := i.state.Get(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, ErrInstanceDoesNotExist
	}

	if instance.Status != state.Running {
		return nil, ErrInstanceIsNotRunning
	}
	opts := container.AttachOptions{
		Stdout: true,
		Stderr: true,
		Stream: true,
	}
	resp, err := i.docker.ContainerAttach(i.ctx, instance.ContainerId, opts)
	if err != nil {
		return nil, err
	}

	logCh := make(chan string)

	go func() {
	READ:
		for {
			select {
			case <-i.ctx.Done():
				break READ
			default:
				l, err := resp.Reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					slog.Error("Received error while reading from attached container logs", "error", err)
				}
				logCh <- l
			}
		}
		close(logCh)
		resp.Close()
	}()

	return logCh, nil
}

func (i *Instances) SendCommand(id uint32, command string) error {
	instance, err := i.state.Get(id)
	if err != nil {
		return err
	}
	if instance == nil {
		return nil
	}

	if instance.Status != state.Running {
		return ErrInstanceIsNotRunning
	}

	opts := container.AttachOptions{
		Stdin: true,
	}
	resp, err := i.docker.ContainerAttach(i.ctx, instance.ContainerId, opts)
	if err != nil {
		return err
	}
	defer resp.Close()

	command, _ = strings.CutSuffix(command, "\n")

	_, err = resp.Conn.Write([]byte(command + "\n"))

	return err
}

// EnsureContainer returns the id of the container for the instance if one exists or otherwise creates a new container.
func (i *Instances) ensureContainer(id uint32) (string, error) {
	instance, err := i.state.Get(id)
	if err != nil {
		return "", err
	}
	if instance == nil {
		return "", ErrInstanceDoesNotExist
	}

	if instance.ContainerId != "" {
		return instance.ContainerId, nil
	}

	// as the instance exists we can safely assume that config is non-nil too
	iconfig, err := i.state.GetConfig(id)
	if err != nil {
		return "", err
	}

	cmd := append([]string{"java", "-jar", iconfig.ServerJar}, iconfig.Args...)

	config := &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: map[nat.Port]struct{}{},
		Tty:          true,
		OpenStdin:    true,
		Cmd:          cmd,
		Image:        "eclipse-temurin:21",
		WorkingDir:   "/var/instance",
	}

	portbinds := make(map[nat.Port][]nat.PortBinding)
	portbinds["25565/tcp"] = []nat.PortBinding{{
		HostPort: strconv.Itoa(iconfig.ListenPort),
	}}

	absWd, err := filepath.Abs(iconfig.DataDirectory)
	if err != nil {
		return "", err
	}
	hostConfig := &container.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:/var/instance", absWd)},
		PortBindings: portbinds,
	}

	resp, err := i.docker.ContainerCreate(i.ctx, config, hostConfig, nil, nil, fmt.Sprintf("instance-%d", id))
	if err != nil {
		return "", err
	}

	for _, warning := range resp.Warnings {
		slog.Warn("Received warning while creating a container", "instanceId", id, "warning", warning)
	}

	err = i.state.SetContainerId(id, resp.ID)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}
