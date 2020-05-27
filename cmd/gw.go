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
	"fmt"
	"github.com/aiziyuer/connectME/util"
	"github.com/cybozu-go/transocks"
	httpDialer "github.com/mwitkow/go-http-dialer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	listenGwPort    int
	listenGwAddress string
)

var gwCmd = &cobra.Command{
	Use:   "gw",
	Short: "work as transparent proxy",
	Long:  `A Gateway to backend proxy.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		util.SetupLogs("/var/log/connectME/gw.log")

		if httpproxy.FromEnvironment().HTTPProxy != "" {
			zap.S().Infof("http_proxy: %s.", httpproxy.FromEnvironment().HTTPProxy)
		}
		if httpproxy.FromEnvironment().HTTPSProxy != "" {
			zap.S().Infof("https_proxy: %s.", httpproxy.FromEnvironment().HTTPSProxy)
		}

		l, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", listenGwAddress, listenGwPort))

		zap.S().Infof("gw_server: %s:%d", listenGwAddress, listenGwPort)

		for {
			conn, _ := l.Accept()

			go func(src net.Conn) {

				origAddr, _ := transocks.GetOriginalDST(src.(*net.TCPConn))

				var dialer *httpDialer.HttpTunnel
				proxyStr := util.GetEnvAny(
					regexp.MustCompile(`http(s)?://`).ReplaceAllString(httpproxy.FromEnvironment().HTTPSProxy, "https://"),
					regexp.MustCompile(`http(s)?://`).ReplaceAllString(httpproxy.FromEnvironment().HTTPProxy, "http://"),
				)

				if strings.TrimSpace(proxyStr) != "" {

					proxyUrl, err := url.Parse(proxyStr)
					if err != nil {
						zap.S().Error(err)
						return
					}
					dialer = httpDialer.New(proxyUrl)

					dest, _ := dialer.Dial("tcp", origAddr.String())

					ch := make(chan error, 2)
					go func() { _, err := io.Copy(src, dest); ioutil.NopCloser(dest); ch <- err }()
					go func() { _, err := io.Copy(dest, src); ioutil.NopCloser(src); ch <- err }()

					for i := 0; i < 2; i++ {
						e := <-ch
						if e != nil {
							zap.S().Error(e)
						}
					}
				}

			}(conn)
		}

	},
}

func init() {
	rootCmd.AddCommand(gwCmd)

	gwCmd.PersistentFlags().IntVar(&listenGwPort, "port", 11081,
		"listen server port",
	)
	gwCmd.PersistentFlags().StringVar(&listenGwAddress, "address", "0.0.0.0",
		"listen server address",
	)
	gwCmd.PersistentFlags().BoolVar(&insecure, "insecure", false,
		"allow insecure server connections when using SSL",
	)
}
