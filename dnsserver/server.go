package dnsserver

import (
	"github.com/aiziyuer/connectDNS/dnsclient"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
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

	// DoH
	doh := dnsclient.NewGoogleDNS(func(option *dnsclient.Option) {
		option.ClientIP = f.option.ClientIP
		option.Client = f.option.Client
	})

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
				doh.LookupRawAppend(r, q.Name, q.Qtype)
			}
		default:
			// dns.TypeTXT, dns.TypeA, dns.TypeAAAA
			doh.LookupRawAppend(r, q.Name, q.Qtype)
		}
	}

	log.Infof("======= query record result ============\n%s", r)
	err := writer.WriteMsg(r)
	if err != nil {
		log.Warnf("Error: Writing Response:%v\n", err)
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
		log.Printf("Error: DNS:" + err.Error())
		return
	}
	// domain not found
	if ret != nil && (ret.Rcode != dns.RcodeSuccess || len(ret.Answer) == 0) {
		resp.SetRcode(resp, dns.RcodeNameError)
		return
	}
	resp.Answer = append(resp.Answer, ret.Answer[0])
}
