package main

import (
	"context"
	"log"

	"github.com/miekg/dns"
)

type DnsServer struct {
}

func NewDns() *DnsServer {
	return &DnsServer{}
}

// func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
// }
