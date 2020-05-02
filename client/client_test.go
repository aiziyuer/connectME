package client

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewGoogleDNS(t *testing.T) {

	client := NewGoogleDNS(func(option *Option) {
		clientIP, _ := GetPublicIP()
		option.ClientIP = clientIP
		option.InsecureSkipVerify = true
	})
	msg := client.Lookup("www.iqiyi.com", dns.TypeA)

	logrus.Info(msg.Answer[0])
}
