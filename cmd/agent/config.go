package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	addressVar   = `ADDRESS`
	reportIntVar = `REPORT_INTERVAL`
	pollIntVar   = `POLL_INTERVAL`
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

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.addr = envAddr
	}

	if envReport := os.Getenv(reportIntVar); envReport != "" {
		reportIntervalEnvVal, err := strconv.Atoi(envReport)
		if err == nil {
			c.reportInterval = reportIntervalEnvVal
		}
	}

	if envPoll := os.Getenv(pollIntVar); envPoll != "" {
		pollIntervalEnvVal, err := strconv.Atoi(envPoll)
		if err == nil {
			c.pollInterval = pollIntervalEnvVal
		} else {
			fmt.Println(err)
		}
	}
}
