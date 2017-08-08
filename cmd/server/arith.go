package main

import (
	"log"
)

type Args struct {
	A int
	B int
}

type Arith int

type Result int

func (t *Arith) Multiply(args *Args, result *Result) error {
	log.Printf("Multiplying %d with %d\n", args.A, args.B)
	*result = Result(args.A * args.B)
	return nil
}
