package main

import (
	"context"
	"log"
	"time"

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

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state.Containers = state.NewContainerList(cancelCtx, cli)
	state.Networks = state.NewNetworkList(cancelCtx, cli)

	ctx := context.WithValue(cancelCtx, "client", cli)

	var timeout chan int

	for {
		var printStatus bool

		select {
		case err := <-state.Containers.Errs:
			log.Fatalf("container error: %v", err)
		case err := <-state.Networks.Errs:
			log.Fatalf("network error: %v", err)
		// case msg := <-state.Containers.Msgs:
		// 	state.Containers.HandleEvent(ctx, msg)
		// 	printStatus = true
		case msg := <-state.Networks.Msgs:
			state.Networks.HandleEvent(ctx, msg)
			printStatus = true
		case <-timeout:
			state.Networks.PrintStatus()
			// state.Containers.PrintStatus()
			timeout = nil
		}

		if printStatus && timeout == nil {
			timeout = make(chan int, 1)

			go func() {
				time.Sleep(5 * time.Second)
				timeout <- 0
			}()
		}
	}
}
