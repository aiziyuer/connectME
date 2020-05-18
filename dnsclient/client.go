package dnsclient

import (
	"fmt"
	"github.com/gogf/gf/encoding/gurl"
	"github.com/miekg/dns"
	"net/http"
	"strings"
	"time"
)

type CustomResolver interface {
	LookupRaw(name string, rType uint16) *dns.Msg
	LookupRawAppend(r *dns.Msg, name string, rType uint16)
	LookupRawA(name string) []*dns.A
	LookupRawTXT(name string) *dns.TXT
}

type Option struct {
	Endpoint string
	Client   *http.Client
	ClientIP string // dnsclient public ip, for cdn
	Hosts    map[string]string
}
type ModOption func(option *Option)

func WithBaseURL(s string) ModOption {
	return func(option *Option) {
		option.Endpoint = s
	}
}

func NewTraditionDNS(modOptions ...ModOption) CustomResolver {

	option := Option{
		Endpoint: "8.8.8.8:53",
		Client: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	return &Tradition{option: &option}

}

func NewCloudFlareDNS(modOptions ...ModOption) CustomResolver {

	return NewDoH(func(option *Option) {

		// default for cloudflare dns
		option.Endpoint = "https://cloudflare-dns.com/dns-query"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

	})
}

func NewGoogleDNS(modOptions ...ModOption) CustomResolver {
	return NewDoH(func(option *Option) {

		// default for google dns
		option.Endpoint = "https://dns.google/resolve"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

		// no need set dnsclient ip
		if len(option.ClientIP) == 0 || strings.Contains(option.Endpoint, "edns_client_subnet=") {
			return
		}

		// try to set dnsclient ip
		if strings.Contains(option.Endpoint, "?") {
			option.Endpoint =
				fmt.Sprintf("%s&edns_client_subnet=%s", option.Endpoint, gurl.Encode(option.ClientIP))
		} else {
			option.Endpoint =
				fmt.Sprintf("%s?edns_client_subnet=%s", option.Endpoint, gurl.Encode(option.ClientIP))
		}

	})
}

func NewDoH(modOptions ...ModOption) CustomResolver {

	option := Option{
		Endpoint: "https://dns.google/resolve",
		Client: &http.Client{
			Timeout:   time.Second,
			Transport: http.DefaultTransport,
		},
		Hosts: map[string]string{
			"dns.google":         "8.8.8.8",
			"cloudflare-dns.com": "1.1.1.1",
		},
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	client := &DoH{option: &option}
	client.RefreshCache()

	return client
}
