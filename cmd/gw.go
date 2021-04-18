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
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/aiziyuer/connectME/util"
	"github.com/avast/retry-go"
	"github.com/cybozu-go/transocks"
	"github.com/gogf/gf/util/gconv"
	httpDialer "github.com/mwitkow/go-http-dialer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"
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

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		done := make(chan bool, 1)

		go func() {
			msg := <-sig
			zap.S().Warnf("receive msg: %s", gconv.String(msg))
			done <- true
		}()

		go func() {

			l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenGwAddress, listenGwPort))
			defer func() {
				_ = l.Close()
			}()
			if err != nil {
				zap.S().Fatal(err)
			}
			zap.S().Infof("gw_server: %s:%d", listenGwAddress, listenGwPort)

			for {
				conn, err := l.Accept()
				if err != nil {
					zap.S().Fatal(err)
				}

				go func(src net.Conn) {
					defer func() {
						ioutil.NopCloser(src)
					}()

					origAddr, err := transocks.GetOriginalDST(conn.(*net.TCPConn))
					if err != nil {
						zap.S().Fatalf("get origAddr error: %s", gconv.String(err))
					}

					var dialer *httpDialer.HttpTunnel
					proxyStr := util.GetAnyString(
						"http://"+regexp.MustCompile(`http(s)?://`).ReplaceAllString(httpproxy.FromEnvironment().HTTPSProxy, ""),
						"http://"+regexp.MustCompile(`http(s)?://`).ReplaceAllString(httpproxy.FromEnvironment().HTTPProxy, ""),
					)

					if strings.TrimSpace(proxyStr) != "" {

						err := retry.Do(
							func() error {

								proxyUrl, err := url.Parse(proxyStr)
								if err != nil {
									zap.S().Error(err)
									return err
								}
								dialer = httpDialer.New(proxyUrl)

								zap.S().Infof("dialer.Dial origAddr(%s) start...", origAddr.String())
								dest, err := dialer.Dial("tcp", origAddr.String())
								if err != nil {
									zap.S().Warnf("dialer.Dial origAddr(%s) with error: %s", origAddr.String(), err)
									return err
								}

								ch := make(chan error, 2)
								go func() { _, err := io.Copy(src, dest); ioutil.NopCloser(dest); ch <- err }()
								go func() { _, err := io.Copy(dest, src); ioutil.NopCloser(src); ch <- err }()

								for i := 0; i < 2; i++ {
									e := <-ch
									if e != nil {
										zap.S().Warnf("origAddr(%s) transfer data failed with error: %s", origAddr.String(), gconv.String(e))
										return e
									}
								}

								return nil
							}, retry.Attempts(5),
						)

						if err != nil {
							zap.S().Fatalf("origAddr(%s) access failed with error: %s which has exceed retry time limit.", origAddr.String(), err)
						}

					}

				}(conn)
			}

		}()

		<-done

		return nil
	},
}

func init() {
	rootCmd.AddCommand(gwCmd)

	gwCmd.PersistentFlags().IntVar(&listenGwPort, "port", 1081,
		"listen server port",
	)
	gwCmd.PersistentFlags().StringVar(&listenGwAddress, "address", "0.0.0.0",
		"listen server address",
	)
	gwCmd.PersistentFlags().BoolVar(&insecure, "insecure", false,
		"allow insecure server connections when using SSL",
	)
}
