package dnsserver

import (
	"github.com/aiziyuer/connectDNS/dnsclient"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
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

	r := new(dns.Msg)
	r.SetReply(msg)
	r.RecursionAvailable = msg.RecursionDesired
	r.Authoritative = true
	r.SetRcode(msg, dns.RcodeSuccess)

	for _, q := range msg.Question {
		switch q.Qtype {
		//default:
		//	DefaultResolver(f.option.Protocol, &q, r)
		case dns.TypePTR:
			//1.0.0.127.in-addr.arpa.
			if dns.Fqdn(q.Name) == "1.0.0.127.in-addr.arpa." {
				r.Answer = append(r.Answer, &dns.PTR{
					Hdr: dns.RR_Header{Name: dns.Fqdn(q.Name), Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 0},
					Ptr: dns.Fqdn(q.Name),
				})
			} else {
				DefaultResolver(f.option.Protocol, &q, r)
			}
		default:
			// dns.TypeTXT, dns.TypeA, dns.TypeAAAA
			// DoH
			doh := dnsclient.NewGoogleDNS(func(option *dnsclient.Option) {
				option.ClientIP = f.option.ClientIP
				option.Client = f.option.Client
			})
			doh.LookupAppend(r, q.Name, q.Qtype)
		}
	}

	logrus.Infof("======= query record result ============\n%s", r)
	err := writer.WriteMsg(r)
	if err != nil {
		logrus.Warnf("Error: Writing Response:%v\n", err)
	}
	_ = writer.Close()

}

func DefaultResolver(protocol string, q *dns.Question, resp *dns.Msg) {

	defaultResolver := &dns.Client{
		Net:          protocol,
		ReadTimeout:  1000 * time.Second,
		WriteTimeout: 1000 * time.Second,
	}

	ret, _, err := defaultResolver.Exchange(
		new(dns.Msg).SetQuestion(q.Name, q.Qtype),
		"114.114.114.114:53",
	)
	// handle failed
	if err != nil {
		resp.SetRcode(resp, dns.RcodeServerFailure)
		logrus.Printf("Error: DNS:" + err.Error())
		return
	}
	// domain not found
	if ret != nil && (ret.Rcode != dns.RcodeSuccess || len(ret.Answer) == 0) {
		resp.SetRcode(resp, dns.RcodeNameError)
		return
	}
	resp.Answer = append(resp.Answer, ret.Answer[0])
}
