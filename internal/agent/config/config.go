package config

import (
	"flag"
	"github.com/sotavant/yandex-metrics/internal"
	"os"
	"strconv"
)

const (
	pollInterval   = 2
	reportInterval = 10
	rateLimit      = 10
	serverAddress  = `localhost:8080`
	addressVar     = `ADDRESS`
	reportIntVar   = `REPORT_INTERVAL`
	pollIntVar     = `POLL_INTERVAL`
	HashKeyVar     = `KEY`
	RateLimitVar   = `RATE_LIMIT`
)

var AppConfig *Config

type Config struct {
	Addr           string
	ReportInterval int
	PollInterval   int
	HashKey        string
	RateLimit      int
}

func InitConfig() {
	AppConfig = new(Config)
	AppConfig.ParseFlags()
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Addr, "a", serverAddress, "server address")
	flag.StringVar(&c.HashKey, "k", "", "hash key")
	flag.IntVar(&c.PollInterval, "p", pollInterval, "pollInterval")
	flag.IntVar(&c.ReportInterval, "r", reportInterval, "reportInterval")
	flag.IntVar(&c.RateLimit, "l", rateLimit, "rate limit")

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
			internal.Logger.Infow("poll interval convert error", "err", err)
		}
	}

	if rateLimit := os.Getenv(RateLimitVar); rateLimit != "" {
		rateLimitEnvVal, err := strconv.Atoi(rateLimit)
		if err == nil {
			c.RateLimit = rateLimitEnvVal
		} else {
			internal.Logger.Infow("rate limit convert error", "err", err)
		}
	}

	if hashKey := os.Getenv(HashKeyVar); hashKey != "" {
		c.HashKey = hashKey
	}
}
