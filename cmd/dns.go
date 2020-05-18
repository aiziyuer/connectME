/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/aiziyuer/connectME/dnsclient"
	"github.com/aiziyuer/connectME/dnsserver"
	"github.com/aiziyuer/connectME/util"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "work as dns server",
	Long:  `A DNS Server support DOH.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		util.SetupLogs("/var/log/connectME/dns.log")

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
			},
		}

		m := dnsclient.NewTraditionDNS(func(option *dnsclient.Option) {
			option.Client = client
		}).LookupRawTXT("o-o.myaddr.l.google.com")
		if m.Txt == nil || len(m.Txt) == 0 {
			zap.S().Fatalf("public_ip can't get!")
		}

		ednsSubnet := ""
		for _, txt := range m.Txt {
			if strings.Contains(txt, "edns") {
				r := regexp.MustCompile(`^edns0-client-subnet (?P<subnet>\S+)$`)
				m := util.NamedStringSubMatch(r, txt)
				if len(m) > 0 {
					ednsSubnet = m["subnet"]
					break
				}
			}
		}
		if len(ednsSubnet) == 0 {
			zap.S().Fatalf("public_ip can't get!")
		}

		zap.S().Infof("ednsSubnet: %s", ednsSubnet)

		if httpproxy.FromEnvironment().HTTPProxy != "" {
			zap.S().Infof("http_proxy: %s.", httpproxy.FromEnvironment().HTTPProxy)
		}
		if httpproxy.FromEnvironment().HTTPSProxy != "" {
			zap.S().Infof("https_proxy: %s.", httpproxy.FromEnvironment().HTTPSProxy)
		}

		// dig @127.0.0.1 -p53 www.google.com A +short
		go func() {
			protocol := "udp"
			h := dns.NewServeMux()
			s := dnsserver.NewForwardServer(func(option *dnsserver.Option) {
				option.ClientIP = ednsSubnet
				option.Protocol = protocol
				option.Client = client
			})
			h.HandleFunc(".", s.Handler)
			zap.S().Infof("%s_server: %s:%d", protocol, listenAddress, listenPort)
			zap.S().Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenAddress, listenPort), protocol, h))
		}()

		// nslookup -vc www.google.com 127.0.0.1
		go func() {
			protocol := "tcp"
			h := dns.NewServeMux()
			s := dnsserver.NewForwardServer(func(option *dnsserver.Option) {
				option.ClientIP = ednsSubnet
				option.Protocol = protocol
				option.Client = client
			})
			h.HandleFunc(".", s.Handler)
			zap.S().Infof("%s_server: %s:%d", protocol, listenAddress, listenPort)
			zap.S().Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenAddress, listenPort), protocol, h))
		}()

		select {}
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	dnsCmd.PersistentFlags().IntVar(&listenPort, "port", 10053,
		"listen server port",
	)
	dnsCmd.PersistentFlags().StringVar(&listenAddress, "address", "0.0.0.0",
		"listen server address",
	)
	dnsCmd.PersistentFlags().BoolVar(&insecure, "insecure", false,
		"allow insecure server connections when using SSL",
	)

}
