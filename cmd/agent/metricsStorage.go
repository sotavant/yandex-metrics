package main

import (
	"math/rand"
	"runtime"
	"time"
)

type MetricsStorage struct {
	Metrics   map[string]float64
	PollCount int64
}

func NewStorage() *MetricsStorage {
	var m MetricsStorage
	m.Metrics = make(map[string]float64)

	return &m
}

func (m *MetricsStorage) updateValues() {
	var rtm runtime.MemStats
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	runtime.ReadMemStats(&rtm)

	m.Metrics["Alloc"] = float64(rtm.Alloc)
	m.Metrics["BuckHashSys"] = float64(rtm.BuckHashSys)
	m.Metrics["Frees"] = float64(rtm.Frees)
	m.Metrics["GCCPUFraction"] = rtm.GCCPUFraction
	m.Metrics["GCSys"] = float64(rtm.GCSys)
	m.Metrics["HeapAlloc"] = float64(rtm.HeapAlloc)
	m.Metrics["HeapIdle"] = float64(rtm.HeapIdle)
	m.Metrics["HeapInuse"] = float64(rtm.HeapInuse)
	m.Metrics["HeapObjects"] = float64(rtm.HeapObjects)
	m.Metrics["HeapReleased"] = float64(rtm.HeapReleased)
	m.Metrics["HeapSys"] = float64(rtm.HeapSys)
	m.Metrics["LastGC"] = float64(rtm.LastGC)
	m.Metrics["Lookups"] = float64(rtm.Lookups)
	m.Metrics["MCacheInuse"] = float64(rtm.MCacheInuse)
	m.Metrics["MCacheSys"] = float64(rtm.MCacheSys)
	m.Metrics["Mallocs"] = float64(rtm.Mallocs)
	m.Metrics["NextGC"] = float64(rtm.NextGC)
	m.Metrics["NumForcedGC"] = float64(rtm.NumForcedGC)
	m.Metrics["NumGC"] = float64(rtm.NumGC)
	m.Metrics["OtherSys"] = float64(rtm.OtherSys)
	m.Metrics["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	m.Metrics["StackInuse"] = float64(rtm.StackInuse)
	m.Metrics["StackSys"] = float64(rtm.StackSys)
	m.Metrics["Sys"] = float64(rtm.Sys)
	m.Metrics["TotalAlloc"] = float64(rtm.TotalAlloc)
	m.Metrics["RandomValue"] = r.Float64()
	m.PollCount += 1
}
