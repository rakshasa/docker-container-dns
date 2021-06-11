package main

import (
	"context"
	"log"
	"net"

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

	ip, ok := writer.RemoteAddr().(*net.UDPAddr)
	if !ok {
		log.Printf("not an udp address")
		return
	}

	log.Printf("request from udp address: %s", ip)

	//var reply dns.RR

	msg := new(dns.Msg)
	msg.SetReply(request)

	if len(request.Question) != 1 {
		log.Printf("unsupported dns question length")
		return
	}
	question := request.Question[0]

	switch question.Qtype {
	case dns.TypeA:
		log.Printf("type A request")

		// reply = &dns.A{
		// 	Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
		// 	A:   a.To4(),
		// }

	case dns.TypeAAAA:
		log.Printf("type AAAA request")
		return
	default:
		log.Printf("unknown dns request type")
		return
	}

	switch question.Qtype {
	case dns.TypeA:
		log.Printf("question A: %v", question.Qclass)
	case dns.TypeAAAA:
		log.Printf("question AAAA: %v", question.Qclass)
	default:
		log.Printf("question unknown: %v", question.Qclass)
	}
}
