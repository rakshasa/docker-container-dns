package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

type ContainerEndpoint struct {
	ContainerID   string
	ContainerName string
	IPv4Address   string
	IPv6Address   string
}

func (e *ContainerEndpoint) String() string {
	v := fmt.Sprintf("id:%s name:%s", e.ContainerID[:12], e.ContainerName)

	if len(e.IPv4Address) != 0 {
		v += " "+e.IPv4Address
	}
	if len(e.IPv6Address) != 0 {
		v += " "+e.IPv6Address
	}

	return v
}

type Network struct {
	ID                 string
	Name               string
	ContainerEndpoints map[string]*ContainerEndpoint
}

func (n *Network) String() string {
	return fmt.Sprintf("id:%s name:%s", n.ID[:12], n.Name)
}

func (n *Network) CompactString() string {
	return fmt.Sprintf("%s:%s", n.ID[:12], n.Name)
}

type networkList struct {
	Networks map[string]*Network
	Msgs     <-chan events.Message
	Errs     <-chan error
}

func NewNetworkList(ctx context.Context) (*networkList, error) {
	filter := filters.NewArgs()
	filter.Add("type", events.NetworkEventType)
	filter.Add("event", "create")
	filter.Add("event", "destroy")
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")

	cli, err := dockerClient(ctx)
	if err != nil {
		return nil, err
	}

	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	return &networkList{
		Networks: make(map[string]*Network),
		Msgs:     msgs,
		Errs:     errs,
	}, nil
}

func (m *networkList) PrintStatus() {
	log.Printf("Networks:")

	for _, nw := range m.Networks {
		log.Printf(" - %v", nw)

		for _, endpoint := range nw.ContainerEndpoints {
			log.Printf("   - %v", endpoint)
		}
	}
}

func (m *networkList) HandleEvent(ctx context.Context, msg events.Message) error {
	if msg.Type != events.NetworkEventType {
		log.Printf("error, not a network event: %v", msg)
		return fmt.Errorf("error, not a network event: %v", msg)
	}

	var err error

	switch msg.Action {
	case "create":
		// log.Printf("network->create: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		err = m.handleCreate(msg)
	case "destroy":
		// log.Printf("network->destroy: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		err = m.handleDestroy(msg)
	case "connect":
		// log.Printf("network->connect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		err = m.handleConnect(ctx, msg)
	case "disconnect":
		// log.Printf("network->disconnect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		err = m.handleDisconnect(msg)
	}

	if err != nil {
		log.Printf("network %s handler error: %v", msg.Action, err)
	}

	return nil
}

func (m *networkList) handleCreate(msg events.Message) error {
	networkID, networkName := msg.Actor.ID, msg.Actor.Attributes["name"]

	if len(networkName) == 0 {
		return fmt.Errorf("message does not contains a valid network name")
	}

	if _, exists := m.Networks[networkID]; exists {
		return fmt.Errorf("skipping already known network: %s", networkName)
	}

	nw := &Network{
		ID:                 networkID,
		Name:               networkName,
		ContainerEndpoints: make(map[string]*ContainerEndpoint),
	}
	m.Networks[networkID] = nw

	log.Printf("added network: %v", nw)
	return nil
}

func (m *networkList) handleDestroy(msg events.Message) error {
	networkID := msg.Actor.ID

	nw, exists := m.Networks[networkID]
	if !exists {
		return nil
	}

	delete(m.Networks, networkID)

	log.Printf("removed network: %v", nw)
	return nil
}

func (m *networkList) handleConnect(ctx context.Context, msg events.Message) error {
	networkID, containerID := msg.Actor.ID, msg.Actor.Attributes["container"]

	containerInspect, networkEndpoint, err := dockerContainerInspectAndNetworkEndpoint(ctx, containerID, networkID)
	if err != nil {
		return fmt.Errorf("could not get container '%s' inspect or endpoint for network '%s': %v", containerInspect.Name, networkID[:12], err)
	}

	nw, exists := m.Networks[networkID]
	if !exists {
		return fmt.Errorf("could not find network: id:%s", networkID[:12])
	}

	if _, exists := nw.ContainerEndpoints[containerID]; exists {
		return fmt.Errorf("container id already exists on network '%s': %s", nw.CompactString(), networkID[:12])
	}

	endpoint := &ContainerEndpoint{
		ContainerID:   containerID,
		ContainerName: containerInspect.Name,
		IPv4Address:   networkEndpoint.IPAddress,
		IPv6Address:   networkEndpoint.GlobalIPv6Address,
	}
	nw.ContainerEndpoints[containerID] = endpoint

	log.Printf("container connected to network '%s': %v", nw.CompactString(), endpoint)
	return nil
}

func (m *networkList) handleDisconnect(msg events.Message) error {
	networkID, containerID := msg.Actor.ID, msg.Actor.Attributes["container"]

	nw, exists := m.Networks[networkID]
	if !exists {
		return fmt.Errorf("could not find network: id:%s", networkID[:12])
	}

	endpoint, exists := nw.ContainerEndpoints[containerID]
	if !exists {
		return fmt.Errorf("container id does not exists on network '%s': %s", nw.CompactString(), networkID[:12])
	}

	delete(nw.ContainerEndpoints, containerID)

	log.Printf("container disconnected from network '%s': %v", nw.CompactString(), endpoint)
	return nil
}
