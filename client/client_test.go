package client

import (
	"crypto/tls"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"
)

func TestNewGoogleDNS(t *testing.T) {

	client := NewGoogleDNS(func(option *Option) {
		clientIP, _ := GetPublicIP()
		option.ClientIP = clientIP
		option.Client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	msg := client.Lookup("www.iqiyi.com", dns.TypeA)

	logrus.Info(msg.Answer[0])
}
