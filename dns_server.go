package main

import (
	"context"
	"log"

	"github.com/miekg/dns"
)

// Add shutdown context.

type DnsServer struct {
	server *dns.Server
}

func NewDnsServer(ctx context.Context) *DnsServer {
	s := &DnsServer{
		server: &dns.Server{
			Addr: ":53",
			Net:  "udp",
		},
	}

	dns.HandleFunc(".", func(writer dns.ResponseWriter, request *dns.Msg) {
		s.handleRequest(ctx, writer, request)
	})

	return s
}

func (s *DnsServer) ListenAndServe() {
	log.Printf("dns server is listening and serving")
	s.server.ListenAndServe()
	log.Printf("dns server finished shutting down")
}

func (s *DnsServer) handleRequest(ctx context.Context, writer dns.ResponseWriter, request *dns.Msg) {
	log.Printf("dns request writer:  %v", writer)
	log.Printf("dns request request: %v", request)
}
