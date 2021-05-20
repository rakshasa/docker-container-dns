package state

import (
	"fmt"
	"log"

	"github.com/docker/docker/api/types/events"
)

type Container struct {
	Name string
}

type containerList struct {
	containers map[string]*Container
}

func NewContainerList() *containerList {
	return &containerList{
		containers: make(map[string]*Container),
	}
}

func (m *containerList) HandleEvent(msg events.Message) error {
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

	if _, exists := m.containers[id]; exists {
		log.Printf("container->create: skipping already known container: %s", name)
		return nil
	}

	log.Printf("container->create: adding container: %s", name)

	m.containers[id] = &Container{
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

	if _, exists := m.containers[id]; !exists {
		log.Printf("container->destroy: skipping unknown container: %s", name)
		return nil
	}

	log.Printf("container->destroy: removing container: %s", name)

	delete(m.containers, id)

	return nil
}
