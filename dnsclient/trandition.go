package dnsclient

import (
	"context"
	"fmt"
	"github.com/gogf/gf/text/gstr"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"net"
)

type Tradition struct {
	option *Option
}

func (c *Tradition) LookupRawAppend(r *dns.Msg, name string, rType uint16) {

	client := &dns.Client{}
	m := new(dns.Msg)
	m.SetQuestion(gstr.ReplaceByMap(fmt.Sprintf("%s.", name), map[string]string{"..": "."}), rType)
	m.RecursionDesired = true
	response, _, err := client.ExchangeContext(context.Background(), m, c.option.Endpoint)
	if err != nil {
		return
	}
	if response.Rcode != dns.RcodeSuccess {
		return
	}

	for _, a := range response.Answer {
		r.Answer = append(r.Answer, a)
	}

}

func (c *Tradition) LookupRawA(name string) []*dns.A {
	panic("implement me")
}

func (c *Tradition) LookupRawTXT(name string) *dns.TXT {

	ctx := context.Background()
	defer ctx.Done()
	r := net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", c.option.Endpoint)
		},
	}

	result, err := r.LookupTXT(ctx, name)
	if err != nil {
		zap.S().Error(err)
	}

	return &dns.TXT{
		Hdr: dns.RR_Header{
			Name:     name,
			Rrtype:   dns.TypeTXT,
			Class:    dns.ClassINET,
			Ttl:      0,
			Rdlength: 0,
		},
		Txt: result,
	}
}

func (c *Tradition) LookupRaw(name string, rType uint16) *dns.Msg {

	ret := new(dns.Msg)
	ret.SetQuestion(name, rType)
	ret.SetRcode(ret, dns.RcodeSuccess)

	c.LookupRawAppend(ret, name, rType)

	return ret
}
