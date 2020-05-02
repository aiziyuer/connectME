package server

import (
	"fmt"
	"github.com/aiziyuer/connectDNS/client"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type Option struct {
	protocol string
	address  string
	port     int
	upstream string
}

type ModOption func(option *Option)

func ForwardServer(modOptions ...ModOption) {

	option := Option{
		protocol: "udp",
		address:  "0.0.0.0",
		port:     53,
		upstream: "",
	}

	for _, fn := range modOptions {
		fn(&option)
	}

	h := dns.NewServeMux()
	h.HandleFunc(".", func(writer dns.ResponseWriter, msg *dns.Msg) {

		r := new(dns.Msg)
		r.SetReply(msg)
		r.RecursionAvailable = msg.RecursionDesired
		r.Authoritative = true
		r.SetRcode(msg, dns.RcodeSuccess)

		for _, q := range msg.Question {
			switch q.Qtype {
			default:
				//defaultResolver(protocol, &q, r)
			case dns.TypePTR:
				//1.0.0.127.in-addr.arpa.
				if dns.Fqdn(q.Name) == "1.0.0.127.in-addr.arpa." {
					r.Answer = append(r.Answer, &dns.PTR{
						Hdr: dns.RR_Header{Name: dns.Fqdn(q.Name), Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 0},
						Ptr: dns.Fqdn(q.Name),
					})
				} else {
					//defaultResolver(protocol, &q, r)
				}
			case dns.TypeA, dns.TypeAAAA:
				// DoH
				doh := client.NewCloudFlareDNS(client.WithBaseURL(option.upstream))
				doh.LookupAppend(r, q.Name, q.Qtype)
			}
		}

		err := writer.WriteMsg(r)
		if err != nil {
			logrus.Warnf("Error: Writing Response:%v\n", err)
		}
		_ = writer.Close()

	})

	logrus.Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", option.address, option.port), option.protocol, h))
}
