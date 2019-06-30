package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/resp"
)

var (
	listenAddr string

	//store
	mapLock  = new(sync.Mutex)
	mapStore = make(map[string]string)
)

func init() {
	flag.StringVar(&listenAddr, "addr", "0.0.0.0:6379", "listen address")
}

func main() {
	config := redeo.Config{
		TCPKeepAlive: 300 * time.Second,
	}
	server := redeo.NewServer(&config)

	registerCommand(server)

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	err = server.Serve(l)
	if err != nil {
		panic(err)
	}
}

/*
PING_INLINE ping
PING_BULK ping
SET set
GET get
INCR incr
LPUSH lpush
RPUSH rpush
LPOP lpop
RPOP rpop
SADD sadd
SPOP spop
LPUSH (needed to benchmark LRANGE) lpush
LRANGE_100 (first 100 elements) lrange
LRANGE_300 (first 300 elements) lrange
LRANGE_500 (first 450 elements) lrange
LRANGE_600 (first 600 elements) lrange
MSET (10 keys) mset
*/
func registerCommand(server *redeo.Server) {
	server.HandleFunc("ping", func(out resp.ResponseWriter, _ *resp.Command) {
		out.AppendInlineString("PONG")
	})
	server.HandleFunc("set", func(out resp.ResponseWriter, req *resp.Command) {
		key := req.Args[0]
		value := req.Args[1]
		mapLock.Lock()
		mapStore[key.String()] = value.String()
		mapLock.Unlock()
		out.AppendInlineString("OK")
	})
	server.HandleFunc("get", func(out resp.ResponseWriter, req *resp.Command) {
		mapLock.Lock()
		v, ok := mapStore[req.Args[0].String()]
		mapLock.Unlock()
		if !ok {
			out.AppendNil()
		} else {
			out.AppendBulkString(v)
		}
	})
	server.HandleFunc("info", func(out resp.ResponseWriter, req *resp.Command) {
		mapLock.Lock()
		msg := fmt.Sprintln("key_count: " + strconv.Itoa(len(mapStore)))
		mapLock.Unlock()
		out.AppendBulkString(msg)
	})
}
