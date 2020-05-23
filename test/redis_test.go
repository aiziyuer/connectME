package test

import (
	"github.com/gogf/gf/encoding/gbinary"
	"github.com/gogf/gf/util/gconv"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/leveldb"
	"github.com/tidwall/redcon"
	"log"
	"net"
	"strings"
	"testing"
)

func TestRedisServer(t *testing.T) {

	GlobalCache, _ := leveldb.NewStore(leveldb.Options{
		Path:      "leveldb",
		Codec:     encoding.JSON,
		WriteSync: true,
	})
	defer func() { _ = GlobalCache.Close() }()

	l, _ := net.Listen("tcp", "0.0.0.0:6379")

	log.Fatal(redcon.Serve(l, func(conn redcon.Conn, cmd redcon.Command) {

		switch strings.ToLower(string(cmd.Args[0])) {
		default:
			conn.WriteError("ERR unknown command '" + gconv.String(cmd.Args[0]) + "'")
		case "select 0":
			conn.WriteString("OK")
			_ = conn.Close()
		case "ping":
			conn.WriteString("PONG")
		case "quit":
			conn.WriteString("OK")
			_ = conn.Close()
		case "set":
			if len(cmd.Args) != 3 {
				conn.WriteError("ERR wrong number of arguments for '" + gconv.String(cmd.Args[0]) + "' command")
				return
			}
			_ = GlobalCache.Set(gconv.String(cmd.Args[1]), gconv.String(cmd.Args[2]))
			conn.WriteString("OK")
		case "get":
			if len(cmd.Args) != 2 {
				conn.WriteError("ERR wrong number of arguments for '" + gconv.String(cmd.Args[0]) + "' command")
				return
			}
			o := new(string)
			ok, _ := GlobalCache.Get(gconv.String(cmd.Args[1]), &o)
			if !ok {
				conn.WriteNull()
			} else {
				conn.WriteBulk(gbinary.EncodeString(*o))
			}
		case "del":
			if len(cmd.Args) != 2 {
				conn.WriteError("ERR wrong number of arguments for '" + gconv.String(cmd.Args[0]) + "' command")
				return
			}
			_ = GlobalCache.Delete(gconv.String(cmd.Args[1]))
		}
	},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	))

}
