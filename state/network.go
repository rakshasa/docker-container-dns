package state

import (
	"log"

	"github.com/docker/docker/api/types/events"
)

type Network struct {
	Name string
}

type networkList struct {
	networks map[string]Network
}

func NewNetworkList() *networkList {
	return &networkList{}
}

func (*networkList) HandleEvent(msg events.Message) {
	switch msg.Action {
	case "create":
		log.Printf("network->create: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
	case "destroy":
		log.Printf("network->destroy: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
	case "connect":
		log.Printf("network->connect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
	case "disconnect":
		log.Printf("network->disconnect: ID:%s %v", msg.Actor.ID, msg.Actor.Attributes)
	// default:
	// 	log.Printf("unknown network message: %s", msg)
	}
}
