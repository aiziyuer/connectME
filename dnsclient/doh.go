package dnsclient

import (
	"bytes"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type DoH struct {
	option *Option
}

func (c *DoH) LookupRaw(name string, rType uint16) *dns.Msg {

	ret := new(dns.Msg)
	ret.SetQuestion(name, rType)
	c.LookupRawAppend(ret, name, rType)

	return ret
}

func (c *DoH) getClient() *resty.Client {
	return resty.
		NewWithClient(c.option.Client).
		SetRetryCount(3).
		SetDebug(false)
}

func (c *DoH) handlerRR(item *DohCommon) (rr dns.RR) {

	switch gconv.Uint16(item.Type) {
	case dns.TypeA:
		rr = &dns.A{
			A: net.ParseIP(item.Data),
		}
	case dns.TypeAAAA:
		rr = &dns.AAAA{
			AAAA: net.ParseIP(item.Data),
		}
	case dns.TypeMX:
		d := strings.Split(item.Data, " ")
		if len(d) < 2 {
			return
		}
		rr = &dns.MX{
			Preference: gconv.Uint16(d[0]),
			Mx:         d[1],
		}
	case dns.TypeNS:
		rr = &dns.NS{
			Ns: item.Data,
		}
	case dns.TypePTR:
		rr = &dns.PTR{
			Ptr: item.Data,
		}
	case dns.TypeCNAME:
		rr = &dns.CNAME{
			Target: item.Data,
		}
	case dns.TypeSOA:
		d := strings.Split(item.Data, " ")
		if len(d) < 7 {
			return
		}
		rr = &dns.SOA{
			Ns:      d[0],
			Mbox:    d[1],
			Serial:  gconv.Uint32(d[2]),
			Refresh: gconv.Uint32(d[3]),
			Retry:   gconv.Uint32(d[4]),
			Expire:  gconv.Uint32(d[5]),
			Minttl:  gconv.Uint32(d[6]),
		}
	case dns.TypeDNSKEY:
	case dns.TypeDS:
	case dns.TypeDLV:
	case dns.TypeSSHFP:
	case dns.TypeNAPTR:
	case dns.TypeSRV:
		d := strings.Split(item.Data, " ")
		if len(d) < 4 {
			return
		}
		rr = &dns.SRV{
			Priority: gconv.Uint16(d[0]),
			Weight:   gconv.Uint16(d[1]),
			Port:     gconv.Uint16(d[2]),
			Target:   d[3],
		}
	case dns.TypeLOC:
	case dns.TypeTXT, dns.TypeSPF, dns.TypeAVC:
		d, err := strconv.Unquote(item.Data)
		if err != nil {
			logrus.Error(err)
			return
		}
		rr = &dns.SPF{
			Txt: []string{d},
		}
	}

	return
}

func (c *DoH) LookupRawAppend(r *dns.Msg, name string, rType uint16) {

	request := c.getClient().R().
		EnableTrace().
		SetHeaders(map[string]string{
			"accept": "application/dns-json",
		}).
		SetQueryParams(map[string]string{
			"name": name,
			"type": dns.TypeToString[rType],
			"cd":   "false", // ignore DNSSEC
			"do":   "false", // ignore DNSSEC
		})

	endpoint := c.option.Endpoint
	if u, err := url.Parse(endpoint); err == nil {
		if ip, ok := c.option.Hosts[u.Host]; ok {
			request.SetHeader("Host", u.Host)
			u.Host = ip
			endpoint = u.String()
		}
	}

	res, err := request.Get(endpoint)
	if err != nil {
		logrus.Fatal(err)
	}

	var resp DohResponse
	if err := json.NewDecoder(bytes.NewReader(res.Body())).Decode(&resp); err != nil {
		logrus.Fatal(err)
	}

	for _, item := range resp.Answer {
		if tmp := c.handlerRR(item); tmp != nil {
			tmp.Header().Name = dns.Fqdn(item.Name)
			tmp.Header().Rrtype = gconv.Uint16(item.Type)
			tmp.Header().Class = dns.ClassINET
			tmp.Header().Ttl = gconv.Uint32(item.TTL)
			r.Answer = append(r.Answer, tmp)
		}
	}

	for _, item := range resp.Authority {
		if tmp := c.handlerRR(item); tmp != nil {
			tmp.Header().Name = dns.Fqdn(item.Name)
			tmp.Header().Rrtype = gconv.Uint16(item.Type)
			tmp.Header().Class = dns.ClassINET
			tmp.Header().Ttl = gconv.Uint32(item.TTL)
			r.Ns = append(r.Ns, tmp)
		}
	}

}

func (c *DoH) LookupRawTXT(name string) *dns.TXT {
	return nil
}

func (c *DoH) LookupRawA(name string) (result []*dns.A) {

	result = make([]*dns.A, 0)

	res, err := resty.NewWithClient(c.option.Client).
		SetDebug(true).
		R().
		EnableTrace().
		SetHeaders(map[string]string{
			"accept": "application/dns-json",
		}).
		SetQueryParams(map[string]string{
			"name": name,
			"type": "A",
			"cd":   "false", // ignore DNSSEC
			"do":   "false", // ignore DNSSEC
		}).
		Get(c.option.Endpoint)
	if err != nil {
		logrus.Fatal(err)
	}

	var resp DohResponse
	if err := json.NewDecoder(bytes.NewReader(res.Body())).Decode(&resp); err != nil {
		logrus.Fatal(err)
	}

	for _, answer := range resp.Answer {

		if gconv.Uint16(answer.Type) != dns.TypeA {
			continue
		}

		result = append(result, &dns.A{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn(answer.Name),
				Rrtype: gconv.Uint16(answer.Type),
				Class:  dns.ClassINET,
				Ttl:    gconv.Uint32(answer.TTL),
			},
			A: net.ParseIP(answer.Data),
		})
	}

	return
}

// requestResponse contains the response from a DNS query.
// Both Google and Cloudflare seem to share a scheme here. As in:
//	https://tools.ietf.org/id/draft-bortzmeyer-dns-json-01.html
//
// https://developers.google.com/speed/public-dns/docs/dns-over-https#dns_response_in_json
// https://developers.cloudflare.com/1.1.1.1/dns-over-https/json-format/
type DohResponse struct {
	Status   int  `json:"Status"` // 0=NOERROR, 2=SERVFAIL - Standard DNS response code (32 bit integer)
	TC       bool `json:"TC"`     // Whether the response is truncated
	RD       bool `json:"RD"`     // Always true for Google Public DNS
	RA       bool `json:"RA"`     // Always true for Google Public DNS
	AD       bool `json:"AD"`     // Whether all response data was validated with DNSSEC
	CD       bool `json:"CD"`     // Whether the dnsclient asked to disable DNSSEC
	Question []struct {
		Name string `json:"name"` // FQDN with trailing dot
		Type int    `json:"type"` // Standard DNS RR type
	} `json:"Question"`
	Answer           []*DohCommon  `json:"Answer"`
	Authority        []*DohCommon  `json:"Authority"`
	Additional       []interface{} `json:"Additional"`
	EdnsClientSubnet string        `json:"edns_client_subnet"` // IP address / scope prefix-length
	Comment          string        `json:"Comment"`            // Diagnostics information in case of an error
}

type DohCommon struct {
	Name string `json:"name"` // Always matches name in the Question section
	Type int    `json:"type"` // Standard DNS RR type
	TTL  int    `json:"TTL"`  // Record's time-to-live in seconds
	Data string `json:"data"` // Data
}
