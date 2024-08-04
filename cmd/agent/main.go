package main

import (
	"log"

	"github.com/brontomc/bronto/agent"
)

func main() {
	err := agent.Start()
	if err != nil {
		log.Fatal(err)
	}
}
