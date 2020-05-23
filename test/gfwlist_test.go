package test

import (
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"log"
	"net"
	"testing"
)

//curl -sfL 'https://dns.google.com/resolve?name=www.iqiyi.com&type=A&edns_client_subnet=115.192.215.114/32' | jq  -r '.Answer[].data'
//curl -sfL 'https://dns.google.com/resolve?name=www.iqiyi.com&type=A' | jq  -r '.Answer[].data'
func TestGFWListParser(t *testing.T) {

	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP("58.215.115.174")
	record, err := db.Country(ip)
	if err != nil {
		log.Fatal(err)
	}

	zap.S().Info(record.Country.IsoCode)
}
