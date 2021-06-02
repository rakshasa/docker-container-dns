package state

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Container struct {
	Name        string
	IPv4Address string
	IPv6Address string
}

type containerList struct {
	Containers map[string]*Container
	Msgs       <-chan events.Message
	Errs       <-chan error
}

func NewContainerList(ctx context.Context, cli *client.Client) *containerList {
	filter := filters.NewArgs()
	filter.Add("type", events.ContainerEventType)
	filter.Add("event", "create")
	filter.Add("event", "destroy")
	filter.Add("event", "start")
	filter.Add("event", "stop")

	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	return &containerList{
		Containers: make(map[string]*Container),
		Msgs:       msgs,
		Errs:       errs,
	}
}

func (m *containerList) PrintStatus() {
	log.Printf("Containers:")

	for id, container := range m.Containers {
		log.Printf(" - %s: %s ipv4:%s ipv6:%s",
			id, container.Name, container.IPv4Address, container.IPv6Address)
	}
}

func (m *containerList) InsertWithMessage(containerId string, networkSettings *network.EndpointSettings) {
	log.Printf("Inserting with message: %v", networkSettings)
}

func (m *containerList) RemoveWithMessage(containerId string, networkSettings *network.EndpointSettings) {
	log.Printf("Removing with message: %v", networkSettings)
}

func (m *containerList) HandleEvent(ctx context.Context, msg events.Message) error {
	if msg.Type != events.ContainerEventType {
		log.Printf("error, not a container event: %v", msg)
		return fmt.Errorf("error, not a container event: %v", msg)
	}

	switch msg.Action {
	case "create":
		log.Printf("container->create: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleCreate(msg)
	case "destroy":
		log.Printf("container->destroy: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return m.handleDestroy(msg)
	case "start":
		log.Printf("container->start: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return nil
	case "stop":
		log.Printf("container->stop: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
		return nil
	default:
		log.Printf("unknown container message: %s", msg)
		return fmt.Errorf("unhandled container event: %v", msg)
	}
}

func (m *containerList) handleCreate(msg events.Message) error{
	id, name := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(id) == 0 {
		return fmt.Errorf("container create event message is missing id: %s", name)
	}
	if len(name) == 0 {
		return fmt.Errorf("container create event message is missing name: %s", name)
	}

	if _, exists := m.Containers[id]; exists {
		log.Printf("container->create: skipping already known container: %s", name)
		return nil
	}

	log.Printf("container->create: adding container: %s", name)

	m.Containers[id] = &Container{
		Name: name,
	}

	return nil
}

func (m *containerList) handleDestroy(msg events.Message) error{
	id, name := msg.Actor.ID, msg.Actor.Attributes["name"]
	if len(id) == 0 {
		return fmt.Errorf("container destroy event message is missing id: %s", name)
	}
	if len(name) == 0 {
		return fmt.Errorf("container destroy event message is missing name: %s", name)
	}

	if _, exists := m.Containers[id]; !exists {
		log.Printf("container->destroy: skipping unknown container: %s", name)
		return nil
	}

	log.Printf("container->destroy: removing container: %s", name)

	delete(m.Containers, id)

	return nil
}
