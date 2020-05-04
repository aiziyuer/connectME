package dnsclient

import (
	"fmt"
	"github.com/miekg/dns"
	"net/http"
	"strings"
	"time"
)

// Client is an interface all clients should conform to.
type Client interface {
	Lookup(name string, rType uint16) *dns.Msg
	LookupAppend(r *dns.Msg, name string, rType uint16)
	LookupA(name string) []*dns.A
	LookupTXT(name string) *dns.TXT
}

type Option struct {
	Endpoint string
	Client   *http.Client
	ClientIP string // dnsclient public ip, for cdn
}
type ModOption func(option *Option)

func WithBaseURL(s string) ModOption {
	return func(option *Option) {
		option.Endpoint = s
	}
}

func NewTraditionDNS(modOptions ...ModOption) Client {

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

func NewCloudFlareDNS(modOptions ...ModOption) Client {

	return NewDoH(func(option *Option) {

		// default for cloudflare dns
		option.Endpoint = "https://1.1.1.1/dns-query"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

	})
}

func NewGoogleDNS(modOptions ...ModOption) Client {
	return NewDoH(func(option *Option) {

		// default for google dns
		option.Endpoint = "https://8.8.8.8/resolve"

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
			option.Endpoint = fmt.Sprintf("%s&edns_client_subnet=%s", option.Endpoint, option.ClientIP)
		} else {
			option.Endpoint = fmt.Sprintf("%s?edns_client_subnet=%s", option.Endpoint, option.ClientIP)
		}

	})
}

func NewDoH(modOptions ...ModOption) Client {

	option := Option{
		Endpoint: "https://8.8.8.8/resolve",
		Client: &http.Client{
			Timeout:   time.Second,
			Transport: http.DefaultTransport,
		},
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	return &DoH{option: &option}
}
