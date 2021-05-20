package main

// go build -mod=readonly -mod=vendor -v
//	"docker.io/go-docker/api/types/filters"

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rakshasa/docker-container-dns/state"
)

const (
	DockerVersion = "1.40"
)

func init() {
}

func main() {
	log.Printf("starting docker-container-dns")

	cli, err := client.NewClientWithOpts(client.WithVersion(DockerVersion))
	if err != nil {
		log.Fatalf("failed to initialize new docker client with version '%s': %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filter := filters.NewArgs()

	filter.Add("type", events.ContainerEventType)
	filter.Add("event", "create")
	filter.Add("event", "destroy")
	filter.Add("event", "start")
	filter.Add("event", "stop")

	filter.Add("type", events.NetworkEventType)
	filter.Add("event", "create")
	filter.Add("event", "destroy")
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")

	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})
	if err != nil {
		log.Fatalf("failed to start listening to docker events: %v", err)
	}

	containers := state.NewContainerList()
	networks := state.NewNetworkList()

	for {
		select {
		case err := <-errs:
			log.Fatalf("read error: %v", err)
		case msg := <-msgs:
			switch msg.Type {
			case events.ContainerEventType:
				containers.HandleEvent(msg)
			case events.NetworkEventType:
				networks.HandleEvent(msg)
			}
		}
	}
}
