package client

import (
	"encoding/json"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type DoH struct {
	option *Option
}

func (c *DoH) Lookup(name string, rType uint16) *dns.Msg {

	ret := new(dns.Msg)
	ret.SetQuestion(name, rType)
	c.LookupAppend(ret, name, rType)

	return ret
}

func (c *DoH) LookupAppend(r *dns.Msg, name string, rType uint16) {

	req, err := http.NewRequest("GET", c.option.Endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("accept", "application/dns-json")

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("type", dns.TypeToString[rType])
	//q.Add("cd", "false") // ignore DNSSEC
	//q.Add("do", "false") // ignore DNSSEC
	req.URL.RawQuery = q.Encode()

	res, err := c.option.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	resp := RequestResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal(err)
	}

	for _, answer := range resp.Answer {

		var rr dns.RR

		switch gconv.Uint16(answer.Type) {
		case dns.TypeA:
			rr = &dns.A{
				Hdr: dns.RR_Header{
					Name:   dns.Fqdn(answer.Name),
					Rrtype: gconv.Uint16(answer.Type),
					Class:  dns.ClassINET,
					Ttl:    gconv.Uint32(answer.TTL),
				},
				A: net.ParseIP(answer.Data),
			}
		case dns.TypeAAAA:
			rr = &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   dns.Fqdn(answer.Name),
					Rrtype: gconv.Uint16(answer.Type),
					Class:  dns.ClassINET,
					Ttl:    gconv.Uint32(answer.TTL),
				},
				AAAA: net.ParseIP(answer.Data),
			}
		}

		if rr != nil {
			r.Answer = append(r.Answer, rr)
		}
	}

}
