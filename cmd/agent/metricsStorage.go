package main

import "runtime"

type MetricsStorage struct {
	Alloc,
	BuckHashSys,
	Frees,
	GCCPUFraction,
	GCSys,
	HeapAlloc,
	HeapIdle,
	HeapInuse,
	HeapObjects,
	HeapReleased,
	HeapSys,
	LastGC,
	Lookups,
	MCacheInuse,
	MCacheSys,
	Mallocs,
	NextGC,
	NumForcedGC,
	NUMGC,
	OtherSys,
	PauseTotalNs,
	StackInuse,
	StackSys,
	Sys,
	TotalAlloc,
	RandomValue float64
	PollCount int64
}

func (m *MetricsStorage) updateValues() {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	m.Alloc = float64(rtm.Alloc)
	m.BuckHashSys = float64(rtm.BuckHashSys)
	m.Frees = float64(rtm.Frees)
	m.GCCPUFraction = rtm.GCCPUFraction
	m.GCSys = float64(rtm.GCSys)
}
