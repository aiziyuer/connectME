package test

import (
	"github.com/aiziyuer/connectDNS/dnsclient"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
	"time"
)

func defaultResolver(protocol string, q *dns.Question, resp *dns.Msg) {

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

// FAQ: https://github.com/d2g/dnsforwarder/blob/master/server.go
func TestUDPDnsForward(t *testing.T) {

	handleFunc := func(protocol string) func(rw dns.ResponseWriter, req *dns.Msg) {
		return func(rw dns.ResponseWriter, req *dns.Msg) {

			r := new(dns.Msg)
			r.SetReply(req)
			r.RecursionAvailable = req.RecursionDesired
			r.Authoritative = true
			r.SetRcode(req, dns.RcodeSuccess)

			for _, q := range req.Question {
				switch q.Qtype {
				default:
					defaultResolver(protocol, &q, r)
				case dns.TypePTR:
					//1.0.0.127.in-addr.arpa.
					if dns.Fqdn(q.Name) == "1.0.0.127.in-addr.arpa." {
						r.Answer = append(r.Answer, &dns.PTR{
							Hdr: dns.RR_Header{Name: dns.Fqdn(q.Name), Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 0},
							Ptr: dns.Fqdn(q.Name),
						})
					} else {
						defaultResolver(protocol, &q, r)
					}
				case dns.TypeA, dns.TypeAAAA:
					// DoH
					doh := dnsclient.NewCloudFlareDNS(dnsclient.WithBaseURL("https://1.1.1.1/dns-query"))
					doh.LookupRawAppend(r, q.Name, q.Qtype)
				}
			}

			err := rw.WriteMsg(r)
			if err != nil {
				logrus.Warnf("Error: Writing Response:%v\n", err)
			}
			_ = rw.Close()

		}
	}

	// dig @127.0.0.1 -p53 www.google.com A +short
	go func() {
		protocol := "udp"
		h := dns.NewServeMux()
		h.HandleFunc(".", handleFunc(protocol))
		log.Fatal(dns.ListenAndServe("0.0.0.0:53", protocol, h))
	}()

	// nslookup -vc www.google.com 127.0.0.1
	go func() {
		protocol := "tcp"
		h := dns.NewServeMux()
		h.HandleFunc(".", handleFunc(protocol))
		log.Fatal(dns.ListenAndServe("0.0.0.0:53", protocol, h))
	}()

	select {}
}
