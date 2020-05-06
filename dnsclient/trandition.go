package dnsclient

import (
	"context"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net"
)

type Tradition struct {
	option *Option
}

func (c *Tradition) LookupRawAppend(r *dns.Msg, name string, rType uint16) {
	panic("implement me")
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
		log.Error(err)
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

	ctx := context.Background()
	defer ctx.Done()
	r := net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", c.option.Endpoint)
		},
	}

	var ips []string
	var err error
	switch gconv.Uint16(rType) {
	case dns.TypeTXT:
		ips, err = r.LookupTXT(ctx, name)
		if err != nil {
			ret.SetRcode(ret, dns.RcodeBadName)
			log.Error(err)
		}
	case dns.TypeA:
		ips, err = r.LookupAddr(ctx, name)
		if err != nil {
			ret.SetRcode(ret, dns.RcodeBadName)
			log.Error(err)
		}
	default:
		ret.SetRcode(ret, dns.RcodeNotImplemented)
	}

	for _, ip := range ips {
		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn(name),
				Rrtype: gconv.Uint16(rType),
				Class:  dns.ClassINET,
			},
			A: net.ParseIP(ip),
		}

		ret.Answer = append(ret.Answer, rr)
	}

	return ret
}
