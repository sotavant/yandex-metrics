package main

import (
	"github.com/sotavant/yandex-metrics/internal"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	serverAddress  = `localhost:8080`
)

var Config = new(config)

func main() {
	Config.parseFlags()

	var poolIntervalDuration = time.Duration(Config.pollInterval) * time.Second
	var reportIntervalDuration = time.Duration(Config.reportInterval) * time.Second
	internal.InitLogger()
	ms := NewStorage()
	forever1 := make(chan bool)
	forever2 := make(chan bool)

	go func() {
		for {
			select {
			case <-forever1:
				return
			default:
				<-time.After(poolIntervalDuration)
				ms.updateValues()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-forever2:
				return
			default:
				<-time.After(reportIntervalDuration)
				reportMetric(ms)
			}
		}
	}()

	<-forever2
	<-forever1
}
