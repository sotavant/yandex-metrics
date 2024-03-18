package main

import (
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/client"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"time"
)

func main() {
	config.InitConfig()

	var poolIntervalDuration = time.Duration(config.AppConfig.PollInterval) * time.Second
	var reportIntervalDuration = time.Duration(config.AppConfig.ReportInterval) * time.Second
	internal.InitLogger()
	ms := storage.NewStorage()
	forever1 := make(chan bool)
	forever2 := make(chan bool)

	go func() {
		for {
			select {
			case <-forever1:
				return
			default:
				<-time.After(poolIntervalDuration)
				ms.UpdateValues()
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
				client.ReportMetric(ms)
			}
		}
	}()

	<-forever2
	<-forever1
}
