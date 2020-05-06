package dnsclient

import (
	"fmt"
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
		option.Endpoint = "https://1.1.1.1/dns-query"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

	})
}

func NewGoogleDNS(modOptions ...ModOption) CustomResolver {
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

func NewDoH(modOptions ...ModOption) CustomResolver {

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
