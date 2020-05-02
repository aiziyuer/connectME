package client

import (
	"crypto/tls"
	"github.com/ipinfo/go-ipinfo/ipinfo"
	"net/http"
)

func GetPublicIP() (string, error) {

	clientIP, err := ipinfo.NewClient(&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}).GetIP(nil)

	if err != nil {
		return "", err
	}

	return clientIP.String(), nil
}
