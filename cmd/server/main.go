package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func main() {
	arith := new(Arith)

	s := rpc.NewServer()
	if err := s.Register(arith); err != nil {
		log.Fatal(err)
	}

	s.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	l, err := net.Listen("tcp", ":4444")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on:", l.Addr())

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
