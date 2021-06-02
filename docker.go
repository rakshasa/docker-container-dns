package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func dockerClient(ctx context.Context) (*client.Client, error) {
	cli, ok := ctx.Value(DockerContextVarName).(*client.Client)
	if !ok {
		return nil, fmt.Errorf("could not get docker client from context")
	}

	return cli, nil
}

func dockerContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	cli, err := dockerClient(ctx)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	containerInspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("could not inspect container: %v", err)
	}

	return containerInspect, nil
}

func dockerContainerInspectAndNetworkEndpoint(ctx context.Context, containerID, networkID string) (types.ContainerJSON, *network.EndpointSettings, error) {
	containerInspect, err := dockerContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, &network.EndpointSettings{}, err
	}

	for _, networkEndpoint := range containerInspect.NetworkSettings.Networks {
		if networkEndpoint.NetworkID == networkID {
			return containerInspect, networkEndpoint, nil
		}
	}

	return types.ContainerJSON{}, &network.EndpointSettings{},
		fmt.Errorf("container is not attached to network: containerID:%s containerName:%s networkID:%s",
			containerID, containerInspect.Name, networkID)
}
