package main

import (
	"github.com/maksadbek/distroleader/agent"
)

func main() {
	agent, _ := agent.New()
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
