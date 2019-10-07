package prometheus

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Up stands up a prom container
func Up(port, dataPath string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", fmt.Errorf("Unable to create docker client: %v", err)
	}

	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: port,
	}

	containerPort, err := nat.NewPort("tcp", "9090")
	if err != nil {
		return "", fmt.Errorf("failed to create port: %v", err)
	}

	portBinding := nat.PortMap{
		containerPort: []nat.PortBinding{hostBinding},
	}
	volumeBinding := []string{
		dataPath + ":/etc/prometheus/data",
	}
	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "prom/prometheus:v2.6.0",
		},
		&container.HostConfig{
			Binds:        volumeBinding,
			PortBindings: portBinding,
		}, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	err = cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	log.Printf("Container %s is started\n", cont.ID)
	return cont.ID, nil
}

// Down takes down a running prom container
func Down(id string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("Unable to create docker client: %v", err)
	}

	cli.ContainerStop(context.Background(), id, nil)
	if err != nil {
		return fmt.Errorf("unable to delete container %s: %v", id, err)
	}

	time.Sleep(10 * time.Second)
	return nil
}
