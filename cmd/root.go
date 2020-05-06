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
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aiziyuer/connectDNS/dnsclient"
	"github.com/aiziyuer/connectDNS/dnsserver"
	"github.com/aiziyuer/connectDNS/regexputil"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/http/httpproxy"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile, listenAddress string
var listenPort int
var insecure bool

var rootCmd = &cobra.Command{
	Use: "connectDNS",
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		localHostMap := map[string]string{
			"dns.google": "8.8.8.8",
		}

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
				DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
					host, port, err := net.SplitHostPort(addr)
					if err != nil {
						return nil, err
					}

					// local cache
					log.Infof("host: %s", host)
					if ip, ok := localHostMap[host]; ok {
						var dialer net.Dialer
						conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
						if err == nil {
							return
						}
					}

					ips, err := net.DefaultResolver.LookupAddr(ctx, host)
					if err != nil {
						return nil, err
					}
					for _, ip := range ips {
						var dialer net.Dialer
						conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
						if err == nil {
							break
						}
					}
					return
				},
			},
		}

		m := dnsclient.NewTraditionDNS(func(option *dnsclient.Option) {
			option.Client = client
		}).LookupRawTXT("o-o.myaddr.l.google.com")
		if m.Txt == nil || len(m.Txt) == 0 {
			log.Fatalf("public_ip can't get!")
		}

		ednsSubnet := ""
		for _, txt := range m.Txt {
			if strings.Contains(txt, "edns") {
				r := regexp.MustCompile(`^edns0-client-subnet (?P<subnet>\S+)$`)
				m := regexputil.NamedStringSubmatch(r, txt)
				if len(m) > 0 {
					ednsSubnet = m["subnet"]
					break
				}
			}
		}
		if len(ednsSubnet) == 0 {
			log.Fatalf("public_ip can't get!")
		}

		log.Infof("ednsSubnet: %s.", ednsSubnet)

		if httpproxy.FromEnvironment().HTTPProxy != "" {
			log.Infof("http_proxy: %s.", httpproxy.FromEnvironment().HTTPProxy)
		}
		if httpproxy.FromEnvironment().HTTPSProxy != "" {
			log.Infof("https_proxy: %s.", httpproxy.FromEnvironment().HTTPSProxy)
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
			log.Infof("%s_server: %s:%d", protocol, listenAddress, listenPort)
			log.Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenAddress, listenPort), protocol, h))
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
			log.Infof("%s_server: %s:%d", protocol, listenAddress, listenPort)
			log.Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenAddress, listenPort), protocol, h))
		}()

		select {}
	}

	rootCmd.PersistentFlags().IntVar(&listenPort, "port", 1053,
		"listen server port",
	)
	rootCmd.PersistentFlags().StringVar(&listenAddress, "address", "0.0.0.0",
		"listen server address",
	)
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false,
		"allow insecure server connections when using SSL",
	)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.connectDNS/config.toml)",
	)

	rootCmd.Version = version
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(path.Join(home, ".connectDNS"))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
