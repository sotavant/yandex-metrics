package main

import (
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	serverAddress  = `localhost:8080`
)

func main() {
	config := new(config)
	config.parseFlags()

	var poolIntervalDuration = time.Duration(config.pollInterval) * time.Second
	var reportIntervalDuration = time.Duration(config.reportInterval) * time.Second
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
				reportMetric(ms, *config)
			}
		}
	}()

	<-forever2
	<-forever1
}
