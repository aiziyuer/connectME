package client

import (
	"fmt"
	"github.com/miekg/dns"
	"strings"
)

// Client is an interface all clients should conform to.
type Client interface {
	Lookup(name string, rType uint16) *dns.Msg
	LookupAppend(r *dns.Msg, name string, rType uint16)
}

type Option struct {
	BaseURL            string
	InsecureSkipVerify bool
	ClientIP           string // client public ip, for cdn
}
type ModOption func(option *Option)

func WithBaseURL(s string) ModOption {
	return func(option *Option) {
		option.BaseURL = s
	}
}

func NewCloudFlareDNS(modOptions ...ModOption) *DoH {

	return NewDoH(func(option *Option) {

		// default for cloudflare dns
		option.BaseURL = "https://1.1.1.1/dns-query"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

	})
}

func NewGoogleDNS(modOptions ...ModOption) *DoH {
	return NewDoH(func(option *Option) {

		// default for google dns
		option.BaseURL = "https://8.8.8.8/resolve"

		// custom
		for _, fn := range modOptions {
			fn(option)
		}

		// no need set client ip
		if len(option.ClientIP) == 0 || strings.Contains(option.BaseURL, "edns_client_subnet=") {
			return
		}

		// try to set client ip
		if strings.Contains(option.BaseURL, "?") {
			option.BaseURL = fmt.Sprintf("%s&edns_client_subnet=%s", option.BaseURL, option.ClientIP)
		} else {
			option.BaseURL = fmt.Sprintf("%s?edns_client_subnet=%s", option.BaseURL, option.ClientIP)
		}

	})
}

func NewDoH(modOptions ...ModOption) *DoH {

	option := Option{
		BaseURL:            "https://8.8.8.8/resolve",
		InsecureSkipVerify: false,
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	return &DoH{Option: &option}
}

// requestResponse contains the response from a DNS query.
// Both Google and Cloudflare seem to share a scheme here. As in:
//	https://tools.ietf.org/id/draft-bortzmeyer-dns-json-01.html
//
// https://developers.google.com/speed/public-dns/docs/dns-over-https#dns_response_in_json
// https://developers.cloudflare.com/1.1.1.1/dns-over-https/json-format/
type RequestResponse struct {
	Status   int  `json:"Status"` // 0=NOERROR, 2=SERVFAIL - Standard DNS response code (32 bit integer)
	TC       bool `json:"TC"`     // Whether the response is truncated
	RD       bool `json:"RD"`     // Always true for Google Public DNS
	RA       bool `json:"RA"`     // Always true for Google Public DNS
	AD       bool `json:"AD"`     // Whether all response data was validated with DNSSEC
	CD       bool `json:"CD"`     // Whether the client asked to disable DNSSEC
	Question []struct {
		Name string `json:"name"` // FQDN with trailing dot
		Type int    `json:"type"` // Standard DNS RR type
	} `json:"Question"`
	Answer []struct {
		Name string `json:"name"` // Always matches name in the Question section
		Type int    `json:"type"` // Standard DNS RR type
		TTL  int    `json:"TTL"`  // Record's time-to-live in seconds
		Data string `json:"data"` // Data
	} `json:"Answer"`
	Additional       []interface{} `json:"Additional"`
	EdnsClientSubnet string        `json:"edns_client_subnet"` // IP address / scope prefix-length
	Comment          string        `json:"Comment"`            // Diagnostics information in case of an error
}
