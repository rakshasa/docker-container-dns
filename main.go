package main

// go build -mod=readonly -mod=vendor -v
//	"docker.io/go-docker/api/types/filters"

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

	msgs, errs := cli.Events(ctx, types.EventsOptions{})
	if err != nil {
		log.Fatalf("failed to start listening to docker events: %v", err)
	}

	for {
		select {
		case err := <-errs:
			log.Fatalf("read error: %v", err)
		case msg := <-msgs:
			log.Printf("read message: %s", msg)
		}
	}
}
