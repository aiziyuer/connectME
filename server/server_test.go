package server

import (
	"crypto/tls"
	"github.com/aiziyuer/connectDNS/client"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"testing"
)

func TestForwardServer(t *testing.T) {

	m := client.NewTraditionDNS().Lookup("o-o.myaddr.l.google.com", dns.TypeTXT)
	if m.Rcode != dns.RcodeSuccess {
		logrus.Fatal("public ip can't found, can't start.")
	}
	result, _ := m.Answer[0].(*dns.A)
	logrus.Info(result.A)

	protocol := "udp"
	h := dns.NewServeMux()
	s := NewForwardServer(func(option *Option) {
		option.clientIP = result.A.String()
		option.protocol = protocol
		option.client = &http.Client{
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
