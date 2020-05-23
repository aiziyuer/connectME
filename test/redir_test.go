package test

import (
	"github.com/cybozu-go/transocks"
	httpDialer "github.com/mwitkow/go-http-dialer"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"testing"
)

// 参考: https://ymmt2005.hatenablog.com/entry/2016/03/13/Transparent_SOCKS_proxy_in_Go_to_replace_NAT
func TestRedir2SocketUpstream(t *testing.T) {

	l, _ := net.Listen("tcp", ":8080")

	for {
		conn, _ := l.Accept()
		go func(src net.Conn) {

			origAddr, _ := transocks.GetOriginalDST(src.(*net.TCPConn))

			// create http tunnel dialer
			proxyUrl, _ := url.Parse("http://10.10.10.254:3128")
			dialer := httpDialer.New(proxyUrl)

			dest, _ := dialer.Dial("tcp", origAddr.String())

			ch := make(chan error, 2)
			go func() { _, err := io.Copy(src, dest); ioutil.NopCloser(dest); ch <- err }()
			go func() { _, err := io.Copy(dest, src); ioutil.NopCloser(src); ch <- err }()

			for i := 0; i < 2; i++ {
				e := <-ch
				if e != nil {
					log.Println(e)
				}
			}

		}(conn)
	}

}
