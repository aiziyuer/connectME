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
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/aiziyuer/connectME/dnsclient"
	"github.com/aiziyuer/connectME/dnsserver"
	"github.com/aiziyuer/connectME/util"
	"github.com/coreos/go-systemd/daemon"
	"github.com/gogf/gf/util/gconv"
	"github.com/miekg/dns"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ednsSubnet string //
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "work as dns server",
	Long:  `A DNS Server support DOH.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		util.SetupLogs("/var/log/connectME/dns.log")
		proxyUri := util.GetEnvAny("proxy") // http://xxxx:3128, socks5://xxxx:1080
		if proxyUri != "" {
			zap.S().Infof("proxy: %s.", proxyUri)
		}

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		done := make(chan bool, 1)

		go func() {
			msg := <-sig
			zap.S().Warnf("receive msg: %s", gconv.String(msg))
			done <- true
		}()

		// 定时上报心跳
		go func() {
			interval, err := daemon.SdWatchdogEnabled(false)
			if err != nil || interval == 0 {
				return
			}
			for {
				daemon.SdNotify(false, daemon.SdNotifyWatchdog)
				time.Sleep(interval / 3)
			}
		}()

		go func() {

			proxy, _ := url.Parse(proxyUri)
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxy),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
					},
				},
			}

			if len(viper.GetString("EDNS_SUBNET")) != 0 {
				ednsSubnet = viper.GetString("EDNS_SUBNET")
			}

			if len(ednsSubnet) == 0 {
				m := dnsclient.NewTraditionDNS(func(option *dnsclient.Option) {
					option.Client = client
				}).LookupRawTXT("o-o.myaddr.l.google.com")
				if m.Txt == nil || len(m.Txt) == 0 {
					zap.S().Fatalf("public_ip can't get!")
				}

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
			}

			zap.S().Infof("ednsSubnet: %s", ednsSubnet)

			proxyUri := util.GetEnvAny("proxy")
			if proxyUri != "" {
				zap.S().Infof("proxy: %s.", proxyUri)
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
				zap.S().Infof("%s_server: %s:%d", protocol, listenDnsAddress, listenDnsPort)
				zap.S().Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenDnsAddress, listenDnsPort), protocol, h))
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
				zap.S().Infof("%s_server: %s:%d", protocol, listenDnsAddress, listenDnsPort)
				zap.S().Fatal(dns.ListenAndServe(fmt.Sprintf("%s:%d", listenDnsAddress, listenDnsPort), protocol, h))
			}()

		}()

		// 正常通知systemd服务已经启动
		daemon.SdNotify(false, daemon.SdNotifyReady)

		<-done

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	dnsCmd.PersistentFlags().IntVar(&listenDnsPort, "port", 53,
		"listen server port",
	)
	dnsCmd.PersistentFlags().StringVar(&listenDnsAddress, "address", "0.0.0.0",
		"listen server address",
	)
	dnsCmd.PersistentFlags().BoolVar(&insecure, "insecure", false,
		"allow insecure server connections when using SSL",
	)

	dnsCmd.PersistentFlags().StringVar(&ednsSubnet, "ednsSubnet", "",
		"ednsSubnet such as 125.119.9.250/32 or env EDNS_SUBNET",
	)

}
