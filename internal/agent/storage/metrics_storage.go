package storage

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const (
	allocMetric         = "Alloc"
	buckHashSysMetric   = "BuckHashSys"
	freesMetric         = "Frees"
	gCCPUFractionMetric = "GCCPUFraction"
	gCSysMetric         = "GCSys"
	heapAllocMetric     = "HeapAlloc"
	heapIdleMetric      = "HeapIdle"
	heapInuseMetric     = "HeapInuse"
	heapObjectsMetric   = "HeapObjects"
	heapReleasedMetric  = "HeapReleased"
	heapSysMetric       = "HeapSys"
	lastGCMetric        = "LastGC"
	lookupsMetric       = "Lookups"
	mCacheInuseMetric   = "MCacheInuse"
	mCacheSysMetric     = "MCacheSys"
	mSpanInUseMetric    = "MSpanInuse"
	mSpanSysMetric      = "MSpanSys"
	mallocsMetric       = "Mallocs"
	nextGCMetric        = "NextGC"
	numForcedGCMetric   = "NumForcedGC"
	numGCMetric         = "NumGC"
	otherSysMetric      = "OtherSys"
	pauseTotalNsMetric  = "PauseTotalNs"
	stackInuseMetric    = "StackInuse"
	stackSysMetric      = "StackSys"
	sysMetric           = "Sys"
	totalAllocMetric    = "TotalAlloc"
	totalMemory         = "TotalMemory"
	freeMemory          = "FreeMemory"
	CPUUtilization1     = "CPUUtilization1"
)

type MetricsStorage struct {
	Metrics   map[string]float64
	PollCount int64
	RWMutex   sync.RWMutex
}

func NewStorage() *MetricsStorage {
	var m MetricsStorage
	m.Metrics = make(map[string]float64)

	return &m
}

func (m *MetricsStorage) UpdateValues() {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()

	var rtm runtime.MemStats
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	runtime.ReadMemStats(&rtm)

	m.Metrics[allocMetric] = float64(rtm.Alloc)
	m.Metrics[buckHashSysMetric] = float64(rtm.BuckHashSys)
	m.Metrics[freesMetric] = float64(rtm.Frees)
	m.Metrics[gCCPUFractionMetric] = rtm.GCCPUFraction
	m.Metrics[gCSysMetric] = float64(rtm.GCSys)
	m.Metrics[heapAllocMetric] = float64(rtm.HeapAlloc)
	m.Metrics[heapIdleMetric] = float64(rtm.HeapIdle)
	m.Metrics[heapInuseMetric] = float64(rtm.HeapInuse)
	m.Metrics[heapObjectsMetric] = float64(rtm.HeapObjects)
	m.Metrics[heapReleasedMetric] = float64(rtm.HeapReleased)
	m.Metrics[heapSysMetric] = float64(rtm.HeapSys)
	m.Metrics[lastGCMetric] = float64(rtm.LastGC)
	m.Metrics[lookupsMetric] = float64(rtm.Lookups)
	m.Metrics[mCacheInuseMetric] = float64(rtm.MCacheInuse)
	m.Metrics[mCacheSysMetric] = float64(rtm.MCacheSys)
	m.Metrics[mSpanInUseMetric] = float64(rtm.MSpanInuse)
	m.Metrics[mSpanSysMetric] = float64(rtm.MSpanSys)
	m.Metrics[mallocsMetric] = float64(rtm.Mallocs)
	m.Metrics[nextGCMetric] = float64(rtm.NextGC)
	m.Metrics[numForcedGCMetric] = float64(rtm.NumForcedGC)
	m.Metrics[numGCMetric] = float64(rtm.NumGC)
	m.Metrics[otherSysMetric] = float64(rtm.OtherSys)
	m.Metrics[pauseTotalNsMetric] = float64(rtm.PauseTotalNs)
	m.Metrics[stackInuseMetric] = float64(rtm.StackInuse)
	m.Metrics[stackSysMetric] = float64(rtm.StackSys)
	m.Metrics[sysMetric] = float64(rtm.Sys)
	m.Metrics[totalAllocMetric] = float64(rtm.TotalAlloc)
	m.Metrics["RandomValue"] = r.Float64()
	m.PollCount += 1
}

func (m *MetricsStorage) UpdateAdditionalValues() {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()

	v, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	cpuTimes, err := cpu.Times(false)
	if err != nil {
		panic(err)
	}

	m.Metrics[totalMemory] = float64(v.Total)
	m.Metrics[freeMemory] = float64(v.Free)
	m.Metrics[CPUUtilization1] = cpuTimes[0].Idle
}
