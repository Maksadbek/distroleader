package main

import (
	"net"
	"net/rpc/jsonrpc"
	"testing"
)

func TestArithRPC(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		t.Fatal(err)
	}

	args := &Args{7, 8}
	var reply int

	c := jsonrpc.NewClient(conn)

	if err = c.Call("Arith.Multiply", args, &reply); err != nil {
		t.Fatal(err)
	}

	t.Logf("arith: %d * %d = %d", args.A, args.B, reply)

}
