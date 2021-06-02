package state

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	Containers *containerList
	Networks   *networkList
)

func dockerContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if len(containerID) == 0 {
		return types.ContainerJSON{}, fmt.Errorf("empty containerID argument")
	}

	cli, ok := ctx.Value("client").(*client.Client)
	if !ok {
		return types.ContainerJSON{}, fmt.Errorf("could not get docker client from context")
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
