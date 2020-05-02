package server

import (
	"github.com/aiziyuer/connectDNS/client"
	"github.com/miekg/dns"
	"log"
	"testing"
)

func TestForwardServer(t *testing.T) {

	protocol := "udp"
	h := dns.NewServeMux()
	clientIP, _ := client.GetPublicIP()
	s := NewForwardServer(func(option *Option) {
		option.clientIP = clientIP
		option.protocol = protocol
		option.insecureSkipVerify = true
	})
	h.HandleFunc(".", s.Handler)

	log.Fatal(dns.ListenAndServe("0.0.0.0:1053", protocol, h))
}
