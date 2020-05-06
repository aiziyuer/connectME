package dnsserver

import (
	"crypto/tls"
	"github.com/aiziyuer/connectDNS/dnsclient"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net/http"
	"testing"
)

func TestForwardServer(t *testing.T) {

	m := dnsclient.NewTraditionDNS().LookupRaw("o-o.myaddr.l.google.com", dns.TypeTXT)
	if m.Rcode != dns.RcodeSuccess {
		log.Fatal("public ip can't found, can't start.")
	}
	result, _ := m.Answer[0].(*dns.A)
	log.Info(result.A)

	protocol := "udp"
	h := dns.NewServeMux()
	s := NewForwardServer(func(option *Option) {
		option.ClientIP = result.A.String()
		option.Protocol = protocol
		option.Client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	h.HandleFunc(".", s.Handler)

	log.Fatal(dns.ListenAndServe("0.0.0.0:1053", protocol, h))
}
