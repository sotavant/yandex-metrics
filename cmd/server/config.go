package main

import (
	"flag"
	"os"
)

const (
	addressVar = `ADDRESS`
)

type config struct {
	addr string
}

func (c *config) parseFlags() {
	flag.StringVar(&c.addr, "a", serverAddress, "server address")
	flag.Parse()

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.addr = envAddr
	}
}
