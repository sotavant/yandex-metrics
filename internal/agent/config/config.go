package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	pollInterval   = 2
	reportInterval = 10
	serverAddress  = `localhost:8080`
	addressVar     = `ADDRESS`
	reportIntVar   = `REPORT_INTERVAL`
	pollIntVar     = `POLL_INTERVAL`
	HashKeyVar     = `KEY`
)

var AppConfig *Config

type Config struct {
	Addr           string
	ReportInterval int
	PollInterval   int
	HashKey        string
}

func InitConfig() {
	AppConfig = new(Config)
	AppConfig.ParseFlags()
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Addr, "a", serverAddress, "server address")
	flag.IntVar(&c.PollInterval, "p", pollInterval, "pollInterval")
	flag.IntVar(&c.ReportInterval, "r", reportInterval, "reportInterval")
	flag.StringVar(&c.HashKey, "k", "", "hash key")

	flag.Parse()

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.Addr = envAddr
	}

	if envReport := os.Getenv(reportIntVar); envReport != "" {
		reportIntervalEnvVal, err := strconv.Atoi(envReport)
		if err == nil {
			c.ReportInterval = reportIntervalEnvVal
		}
	}

	if envPoll := os.Getenv(pollIntVar); envPoll != "" {
		pollIntervalEnvVal, err := strconv.Atoi(envPoll)
		if err == nil {
			c.PollInterval = pollIntervalEnvVal
		} else {
			fmt.Println(err)
		}
	}

	if hashKey := os.Getenv(HashKeyVar); hashKey != "" {
		c.HashKey = hashKey
	}
}
