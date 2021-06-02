package state

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type ContainerEndpoint struct {
	ContainerID   string
	ContainerName string
	IPv4Address   string
	IPv6Address   string
}

type Network struct {
	Name               string
	ContainerEndpoints map[string]*ContainerEndpoint
}

type networkList struct {
	Networks map[string]*Network
	Msgs     <-chan events.Message
	Errs     <-chan error
}

func NewNetworkList(ctx context.Context, cli *client.Client) *networkList {
	filter := filters.NewArgs()
	filter.Add("type", events.NetworkEventType)
	filter.Add("event", "create")
	filter.Add("event", "destroy")
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")

	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	return &networkList{
		Networks: make(map[string]*Network),
		Msgs:     msgs,
		Errs:     errs,
	}
}

func (m *networkList) PrintStatus() {
	log.Printf("Networks:")

	for id, nw := range m.Networks {
		log.Printf(" - %s: %s", id, nw.Name)

		for containerID, endpoint := range nw.ContainerEndpoints {
			log.Printf("   - %s: %s", containerID, endpoint.ContainerName)

			if len(endpoint.IPv4Address) != 0 {
				log.Printf("     %65s: %s", "", endpoint.IPv4Address)
			}
			if len(endpoint.IPv6Address) != 0 {
				log.Printf("     %65s: %s", "", endpoint.IPv6Address)
			}
		}
	}
}

func (m *networkList) HandleEvent(ctx context.Context, msg events.Message) error {
	if msg.Type != events.NetworkEventType {
		log.Printf("error, not a network event: %v", msg)
		return fmt.Errorf("error, not a network event: %v", msg)
	}

	switch msg.Action {
	case "create":
		log.Printf("network->create: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleCreate(msg)
	case "destroy":
		log.Printf("network->destroy: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleDestroy(msg)
	case "connect":
		log.Printf("network->connect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleConnect(ctx, msg)
	case "disconnect":
		log.Printf("network->disconnect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleDisconnect(msg)
	default:
		log.Printf("unknown network message: %s", msg)
		return fmt.Errorf("unhandled network event: %v", msg)
	}
}

func (m *networkList) handleCreate(msg events.Message) error{
	networkID, networkName := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(networkID) == 0 {
		return fmt.Errorf("network create event message is missing id attribute")
	}
	if len(networkName) == 0 {
		return fmt.Errorf("network create event message is missing name attribute")
	}

	if _, exists := m.Networks[networkID]; exists {
		log.Printf("network->create: skipping already known network: %s", networkName)
		return nil
	}

	log.Printf("network->create: adding network: %s", networkName)

	m.Networks[networkID] = &Network{
		Name:               networkName,
		ContainerEndpoints: make(map[string]*ContainerEndpoint),
	}

	return nil
}

func (m *networkList) handleDestroy(msg events.Message) error{
	networkID, networkName := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(networkID) == 0 {
		return fmt.Errorf("network destroy event message is missing id attribute")
	}
	if len(networkName) == 0 {
		return fmt.Errorf("network destroy event message is missing name attribute")
	}

	if _, exists := m.Networks[networkID]; !exists {
		log.Printf("network->destroy: skipping unknown network: %s", networkName)
		return nil
	}

	log.Printf("network->destroy: removing network: %s", networkName)

	delete(m.Networks, networkID)

	return nil
}

func (m *networkList) handleConnect(ctx context.Context, msg events.Message) error {
	networkID, networkName, containerID := msg.Actor.ID, msg.Actor.Attributes["name"], msg.Actor.Attributes["container"]

	log.Printf("container connected to network: containerID:%s networkID:%s networkName:%s",
		containerID, networkID, networkName)

	containerInspect, networkEndpoint, err := dockerContainerInspectAndNetworkEndpoint(ctx, containerID, networkID)
	if err != nil {
		return fmt.Errorf("could not get container '%s' inspect or network '%s' inspect: %v", containerInspect.Name, networkName, err)
	}

	nw, exists := m.Networks[networkID]
	if !exists {
		return fmt.Errorf("could not find network: networkID:%s networkName:%s", networkID, networkName)
	}

	if _, exists := nw.ContainerEndpoints[containerID]; exists {
		return fmt.Errorf("container already exists on network: containerID:%s networkID:%s networkName:%s",
			containerID, networkID, networkName)
	}

	nw.ContainerEndpoints[containerID] = &ContainerEndpoint{
		ContainerID:   containerID,
		ContainerName: containerInspect.Name,
		IPv4Address:   networkEndpoint.IPAddress,
		IPv6Address:   networkEndpoint.GlobalIPv6Address,
	}

	// Print ip addresses:

	return nil
}

func (m *networkList) handleDisconnect(msg events.Message) error {
	networkID, networkName, containerID := msg.Actor.ID, msg.Actor.Attributes["name"], msg.Actor.Attributes["container"]

	log.Printf("network->disconnect: %s:%s -> %s", networkID, networkName, containerID)

	// Containers.RemoveWithMessage(containerID, networkEndpoint)


	// Containers.RemoveWithMessage(containerInspect.Name, networkEndpoint)

	return nil
}
