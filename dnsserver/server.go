package dnsserver

import (
	"github.com/aiziyuer/connectME/dnsclient"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"net/http"
)

type Option struct {
	Protocol string
	upstream string
	ClientIP string
	Client   *http.Client
}

type ModOption func(option *Option)

type ForwardServer struct {
	option *Option
}

func NewForwardServer(modOptions ...ModOption) *ForwardServer {

	option := Option{
		Protocol: "udp",
		upstream: "",
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	return &ForwardServer{
		option: &option,
	}
}

func (f *ForwardServer) Handler(writer dns.ResponseWriter, msg *dns.Msg) {

	defer func() { _ = writer.Close() }()

	r := new(dns.Msg)
	r.SetReply(msg)
	r.RecursionAvailable = msg.RecursionDesired
	r.Authoritative = true
	r.SetRcode(msg, dns.RcodeSuccess)

	// DoH
	doh := dnsclient.NewGoogleDNS(func(option *dnsclient.Option) {
		option.ClientIP = f.option.ClientIP
		option.Client = f.option.Client
	})

	for _, q := range msg.Question {
		switch q.Qtype {
		//default:
		//	DefaultResolver(f.option.Protocol, &q, r)
		case dns.TypePTR:
			//1.0.0.127.in-addr.arpa.
			if dns.Fqdn(q.Name) == "1.0.0.127.in-addr.arpa." || gstr.HasSuffix(dns.Fqdn(q.Name), "10.10.10.in-addr.arpa.") {
				r.Answer = append(r.Answer, &dns.PTR{
					Hdr: dns.RR_Header{Name: dns.Fqdn(q.Name), Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 0},
					Ptr: dns.Fqdn(q.Name),
				})
			} else {
				doh.LookupRawAppend(r, q.Name, q.Qtype)
			}
		default:
			// dns.TypeTXT, dns.TypeA, dns.TypeAAAA
			doh.LookupRawAppend(r, q.Name, q.Qtype)
		}
	}

	if len(r.Answer) == 0 && len(r.Ns) == 0 {
		zap.S().Warn(gconv.String(r))
	} else {
		zap.S().Debug(gconv.String(r))
	}

	if err := writer.WriteMsg(r); err != nil {
		zap.S().Warnf("Error: Writing Response:%v\n", err)
	}
}
