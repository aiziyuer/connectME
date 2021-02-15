package dnsclient

import (
	"crypto/tls"
	"fmt"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestNewGoogleDNS(t *testing.T) {

	client := NewGoogleDNS(func(option *Option) {
		option.ClientIP = "60.186.195.38/32"
		option.Client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	msg := client.LookupRawA("www.iqiyi.com")

	log.Info(msg)
}

func TestTradition_LookupRaw(t *testing.T) {

	c := &dns.Client{}

	m := new(dns.Msg)
	m.SetQuestion("ddns-checkip.quickconnect.to", dns.TypeA)
	m.RecursionDesired = true
	r, _, err := c.ExchangeContext(context.Background(), m, "114.114.114.114:53")
	if err != nil {
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		return
	}
	for _, a := range r.Answer {
		if mx, ok := a.(*dns.MX); ok {
			fmt.Printf("%s\n", mx.String())
		}
	}

}
