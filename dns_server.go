package main

import (
	"context"
	"log"
	"net"
	"strings"

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
		log.Printf("dns_request: not an udp address")
		return
	}

	log.Printf("dns_request: request from udp address: %s", ip)

	//var reply dns.RR

	msg := new(dns.Msg)
	msg.SetReply(request)

	if len(request.Question) != 1 {
		log.Printf("dns_request: unsupported dns question length")
		return
	}
	question := request.Question[0]

	networkList, ok := ctx.Value(NetworksVarName).(*networkList)
	if !ok {
		log.Printf("dns_request: could not get networks list from context")
	}

	var queryAddress string

	switch question.Qtype {
	case dns.TypeA:
		log.Printf("question A (%d): %s", question.Qclass, question.Name)

		trimmedName := strings.TrimSuffix(question.Name, ".rt.")

		if len(trimmedName) != len(question.Name) {
			queryAddress = trimmedName
		}

		if len(queryAddress) != 0 {
			endpoint := networkList.QueryEndpoint(queryAddress)

			log.Printf("dns_request: matched '%s': %v", queryAddress, endpoint)

			rr := &dns.A{
				Hdr: dns.RR_Header{
					Name: question.Name,
					Rrtype: dns.TypeA,
					Class: dns.ClassINET,
					Ttl: 0,
				},
				A: net.ParseIP(endpoint.IPv4Address),
			}

			log.Printf("dns_request: RR: %v", rr)

			msg.Answer = append(msg.Answer, rr)
			// m.Extra = append(m.Extra, t)
		}

	case dns.TypeAAAA:
		log.Printf("question AAAA (%d): %s", question.Qclass, question.Name)
	default:
		log.Printf("question unknown (%d): %s", question.Qclass, question.Name)
	}

	log.Printf("dns_request: write message: %v", msg)

	writer.WriteMsg(msg)
}
