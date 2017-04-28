package main

import (
	"flag"
	"time"
	"sync"

	"github.com/bsm/redeo"
	"fmt"
	"strconv"
)

var (
	listenAddr string

	//store
	mapLock = new(sync.Mutex)
	mapStore = make(map[string]string)
)

func init() {
	flag.StringVar(&listenAddr, "addr", "0.0.0.0:6379", "listen address")
}

func main() {
	config := redeo.Config{
		Addr:listenAddr,
		TCPKeepAlive:300*time.Second,
	}
	server := redeo.NewServer(&config)

	registerCommand(server)

	err := server.ListenAndServe()
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
	server.HandleFunc("ping", func(out *redeo.Responder, _ *redeo.Request) error {
		out.WriteInlineString("PONG")
		return nil
	})
	server.HandleFunc("set", func(out *redeo.Responder, req *redeo.Request) error {
		key := req.Args[0]
		value := req.Args[1]
		mapLock.Lock()
		mapStore[key] = value
		mapLock.Unlock()
		out.WriteInlineString("OK")
		return nil
	})
	server.HandleFunc("get", func(out *redeo.Responder, req *redeo.Request) error {
		mapLock.Lock()
		v, ok:=mapStore[req.Args[0]]
		mapLock.Unlock()
		if !ok {
			out.WriteNil()
		} else {
			out.WriteString(v)
		}
		return nil
	})
	server.HandleFunc("info", func(out *redeo.Responder, req *redeo.Request) error {
		mapLock.Lock()
		msg := fmt.Sprintln("key_count: "+strconv.Itoa(len(mapStore)))
		mapLock.Unlock()
		out.WriteString(msg)
		return nil
	})
}
