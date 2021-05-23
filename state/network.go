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

type Network struct {
	Name string
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

	for id, network := range m.Networks {
		log.Printf(" - %s: %s", id, network.Name)
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
		return m.handleDestroy(msg)
		return nil
	case "disconnect":
		log.Printf("network->disconnect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return nil
	default:
		log.Printf("unknown network message: %s", msg)
		return fmt.Errorf("unhandled network event: %v", msg)
	}
}

func (m *networkList) handleCreate(msg events.Message) error{
	id, name := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(id) == 0 {
		return fmt.Errorf("network create event message is missing id attribute")
	}
	if len(name) == 0 {
		return fmt.Errorf("network create event message is missing name attribute")
	}

	if _, exists := m.Networks[id]; exists {
		log.Printf("network->create: skipping already known network: %s", name)
		return nil
	}

	log.Printf("network->create: adding network: %s", name)

	m.Networks[id] = &Network{
		Name: name,
	}

	return nil
}

func (m *networkList) handleDestroy(msg events.Message) error{
	id, name := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(id) == 0 {
		return fmt.Errorf("network destroy event message is missing id attribute")
	}
	if len(name) == 0 {
		return fmt.Errorf("network destroy event message is missing name attribute")
	}

	if _, exists := m.Networks[id]; !exists {
		log.Printf("network->destroy: skipping unknown network: %s", name)
		return nil
	}

	log.Printf("network->destroy: removing network: %s", name)

	delete(m.Networks, id)

	return nil
}

func (m *networkList) handleConnect(msg events.Message) error{
	id, name, container := msg.Actor.ID, msg.Actor.Attributes["name"], msg.Actor.Attributes["container"]
	if len(id) == 0 {
		return fmt.Errorf("network connect event message is missing id attribute")
	}
	if len(name) == 0 {
		return fmt.Errorf("network connect event message is missing name attribute")
	}
	if len(container) == 0 {
		return fmt.Errorf("network connect event message is missing container attribute")
	}

	log.Printf("network->connect: adding network: %s:%s -> %s", id, name, container)

	return nil
}
