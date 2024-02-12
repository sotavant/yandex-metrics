package main

import (
	"flag"
)

type config struct {
	addr string
}

func (c *config) parseFlags() {
	flag.StringVar(&c.addr, "a", serverAddress, "server address")
	flag.Parse()
}
