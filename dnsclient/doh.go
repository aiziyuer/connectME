package dnsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aiziyuer/connectME/util"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/encoding/gparser"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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
			log.Error(err)
			return
		}
		rr = &dns.SPF{
			Txt: []string{d},
		}
	}

	return
}

var (
	globalCache = cache.New(5*time.Minute, 1*time.Second)
	once        sync.Once
)

func (c *DoH) RefreshCache() {

	once.Do(func() {

		// add go routine to refresh cache
		interval := 10 * 1000 * time.Millisecond
		gtimer.Add(interval, func() {
			now := time.Now()
			for key, item := range globalCache.Items() {
				// Refresh at 15s before expiration
				if time.Duration(item.Expiration-now.UnixNano()) < 15*1000*time.Millisecond {
					m := util.NamedStringSubMatch(regexp.MustCompile(`(?P<name>.+)->(?P<type>.+)`), key)
					if len(m) == 2 {
						var dohResponse DohResponse
						c.lookUP(m["name"], m["type"], &dohResponse)
					}
				}

			}
		})

	})

}

func (c *DoH) LookupRawAppend(r *dns.Msg, name string, rType uint16) {

	cachedKey := fmt.Sprintf("%s->%s", name, dns.TypeToString[rType])
	var dohResponse DohResponse

	// Try cache
	if cachedValue, found := globalCache.Get(cachedKey); found {
		if err := gjson.DecodeTo(cachedValue, &dohResponse); err != nil {
			zap.S().Error(err)
		} else {
			zap.S().Debugf("cachedKey: %s, cachedValue: %s.\n", cachedKey, cachedValue)
		}
	} else {
		if c.lookUP(name, dns.TypeToString[rType], &dohResponse) {
			return
		}
	}

	for _, item := range dohResponse.Answer {
		if tmp := c.handlerRR(item); tmp != nil {
			tmp.Header().Name = dns.Fqdn(item.Name)
			tmp.Header().Rrtype = gconv.Uint16(item.Type)
			tmp.Header().Class = dns.ClassINET
			tmp.Header().Ttl = gconv.Uint32(item.TTL)
			r.Answer = append(r.Answer, tmp)
		}
	}

	for _, item := range dohResponse.Authority {
		if tmp := c.handlerRR(item); tmp != nil {
			tmp.Header().Name = dns.Fqdn(item.Name)
			tmp.Header().Rrtype = gconv.Uint16(item.Type)
			tmp.Header().Class = dns.ClassINET
			tmp.Header().Ttl = gconv.Uint32(item.TTL)
			r.Ns = append(r.Ns, tmp)
		}
	}

}

func (c *DoH) lookUP(name string, rType string, dohResponse *DohResponse) bool {

	cachedKey := fmt.Sprintf("%s->%s", name, rType)

	request := util.NewRequest(c.option.Client).
		SetHeaders(map[string]string{
			"accept": "application/dns-json",
		}).
		SetQueryParams(map[string]string{
			"name": name,
			"type": rType,
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
		zap.S().Error(err)
		return true
	}

	if err := json.NewDecoder(bytes.NewReader(res.Body())).Decode(dohResponse); err != nil {
		zap.S().Error(err)
		return true
	}
	// find the shortest ttl
	ttl := 60 * 60 * 24 // one day
	for _, item := range dohResponse.Answer {
		if item.TTL < ttl {
			ttl = item.TTL
		}
	}
	for _, item := range dohResponse.Authority {
		if item.TTL < ttl {
			ttl = item.TTL
		}
	}

	cacheValue := gparser.MustToJsonString(dohResponse)
	cacheTTL := time.Duration(ttl) * time.Second
	globalCache.Set(cachedKey, cacheValue, cacheTTL)

	return false
}

func (c *DoH) LookupRawTXT(name string) *dns.TXT {
	return nil
}

func (c *DoH) LookupRawA(name string) (result []*dns.A) {

	result = make([]*dns.A, 0)

	res, err := util.NewRequest(c.option.Client).
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
		log.Fatal(err)
	}

	var resp DohResponse
	if err := json.NewDecoder(bytes.NewReader(res.Body())).Decode(&resp); err != nil {
		log.Fatal(err)
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
