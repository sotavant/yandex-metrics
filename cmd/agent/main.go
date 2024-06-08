package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/client"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	internal.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	internal.InitLogger()
	config.InitConfig()

	var poolIntervalDuration = time.Duration(config.AppConfig.PollInterval) * time.Second
	var reportIntervalDuration = time.Duration(config.AppConfig.ReportInterval) * time.Second
	ms := storage.NewStorage()
	updateValuesChan := make(chan bool)
	reportMetricsChan := make(chan bool)
	updateAddValuesChan := make(chan bool)
	pprofChan := make(chan bool)

	go func() {
		err := http.ListenAndServe(":8081", nil)

		if err != nil {
			close(pprofChan)
			panic(err)
		}
	}()

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
	<-pprofChan
}
