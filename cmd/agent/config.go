package main

import (
	"flag"
)

type config struct {
	addr           string
	reportInterval int
	pollInterval   int
}

func (c *config) parseFlags() {
	flag.StringVar(&c.addr, "a", serverAddress, "server address")
	flag.IntVar(&c.pollInterval, "p", pollInterval, "pollInterval")
	flag.IntVar(&c.reportInterval, "r", reportInterval, "reportInterval")
	flag.Parse()
}
