package main

import (
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/client"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"time"
)

func main() {
	internal.InitLogger()
	config.InitConfig()

	var poolIntervalDuration = time.Duration(config.AppConfig.PollInterval) * time.Second
	var reportIntervalDuration = time.Duration(config.AppConfig.ReportInterval) * time.Second
	ms := storage.NewStorage()
	updateValuesChan := make(chan bool)
	reportMetricsChan := make(chan bool)
	updateAddValuesChan := make(chan bool)

	go func() {
		for {
			select {
			case <-updateAddValuesChan:
				return
			default:
				<-time.After(poolIntervalDuration)
				ms.UpdateAdditionalValues()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-updateValuesChan:
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
			case <-reportMetricsChan:
				return
			default:
				<-time.After(reportIntervalDuration)
				client.ReportMetric(ms, config.AppConfig.RateLimit)
			}
		}
	}()

	<-reportMetricsChan
	<-updateValuesChan
	<-updateAddValuesChan
}
