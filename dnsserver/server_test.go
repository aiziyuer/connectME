package dnsserver

import (
	"crypto/tls"
	"github.com/aiziyuer/connectME/dnsclient"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func TestForwardServer(t *testing.T) {

	m := dnsclient.NewTraditionDNS().LookupRaw("o-o.myaddr.l.google.com", dns.TypeTXT)
	if m.Rcode != dns.RcodeSuccess {
		zap.S().Fatal("public ip can't found, can't start.")
	}
	result, _ := m.Answer[0].(*dns.TXT)
	zap.S().Info(result.Txt)

	protocol := "udp"
	h := dns.NewServeMux()
	s := NewForwardServer(func(option *Option) {
		option.ClientIP = result.Txt[0]
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

	zap.S().Fatal(dns.ListenAndServe("0.0.0.0:1053", protocol, h))
}
