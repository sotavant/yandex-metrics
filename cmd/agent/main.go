package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/client"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"github.com/sotavant/yandex-metrics/internal/utils"
)

// Build info.
// Need define throw ldflags:
//
//	go build -ldflags "-X main.buildVersion=0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'"
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	internal.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	internal.InitLogger()
	config.InitConfig()

	var poolIntervalDuration = time.Duration(config.AppConfig.PollInterval) * time.Second
	var reportIntervalDuration = time.Duration(config.AppConfig.ReportInterval) * time.Second
	var err error

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ms := storage.NewStorage()
	ch, err := utils.NewCipher("", config.AppConfig.CryptoKeyPath)
	if err != nil {
		panic(err)
	}

	r := client.NewReporter(ch)

	updateValuesChan := make(chan bool)
	reportMetricsChan := make(chan bool)
	updateAddValuesChan := make(chan bool)
	pprofChan := make(chan bool)

	go func() {
		err = http.ListenAndServe(":8082", nil)

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
				shutdown := r.ReportMetric(ms, config.AppConfig.RateLimit, sigs)
				if shutdown {
					close(pprofChan)
					close(updateAddValuesChan)
					close(updateValuesChan)
					close(reportMetricsChan)
				}
			}
		}
	}()

	<-reportMetricsChan
	<-updateValuesChan
	<-updateAddValuesChan
	<-pprofChan
}
