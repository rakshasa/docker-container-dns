package main

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/client"
)

const (
	DockerVersion = "1.40"
	DockerContextVarName = "docker-client"
)

func init() {
}

func newContextAndCancel() (context.Context, context.CancelFunc) {
	cli, err := client.NewClientWithOpts(client.WithVersion(DockerVersion))
	if err != nil {
		log.Fatalf("failed to initialize new docker client with version '%s': %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, DockerContextVarName, cli)

	return ctx, cancel
}

func main() {
	log.Printf("starting docker-container-dns")

	ctx, cancel := newContextAndCancel()
	defer cancel()

	networks, err := NewNetworkList(ctx)
	if err != nil {
		log.Fatalf("failed to create docker network list: %v", err)
	}
	if err := networks.LoadList(ctx); err != nil {
		log.Fatalf("failed to load docker network list: %v", err)
	}

	go func() {
		dnsServer := NewDnsServer(ctx)
		dnsServer.ListenAndServe()
	}()

	var timeout chan int

	for {
		var printStatus bool

		select {
		case err := <-networks.Errs:
			log.Fatalf("network error: %v", err)
		case msg := <-networks.Msgs:
			if err := networks.HandleEvent(ctx, msg); err != nil {
				log.Printf("unhandled network message error: %v", err)
			}

			printStatus = true
		case <-timeout:
			networks.PrintStatus()
			timeout = nil
		}

		if printStatus && timeout == nil {
			timeout = make(chan int, 1)

			go func() {
				time.Sleep(30 * time.Second)
				timeout <- 0
			}()
		}
	}
}
