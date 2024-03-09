package main

import (
	"fmt"
	"github.com/sotavant/yandex-metrics/internal"
)

type Storage interface {
	AddGaugeValue(key string, value float64)
	AddCounterValue(key string, value int64)
	GetValue(mType string, key string) interface{}
	GetGauge() map[string]float64
	GetCounters() map[string]int64
	GetCounterValue(key string) int64
	GetGaugeValue(key string) float64
	KeyExist(mType string, key string) bool
	AddValue(m internal.Metrics) error
}

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func (m *MemStorage) AddGaugeValue(key string, value float64) {
	m.Gauge[key] = value
}

func (m *MemStorage) AddCounterValue(key string, value int64) {
	m.Counter[key] += value
}

func (m *MemStorage) AddValue(metric internal.Metrics) error {
	switch metric.MType {
	case gaugeType:
		m.AddGaugeValue(metric.ID, *metric.Value)
	case counterType:
		m.AddCounterValue(metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("undefinde type: %s", metric.MType)
	}

	return nil
}

func (m *MemStorage) GetValue(mType, key string) interface{} {
	switch mType {
	case gaugeType:
		val, ok := m.Gauge[key]
		if ok {
			return val
		}
	case counterType:
		val, ok := m.Counter[key]
		if ok {
			return val
		}
	}

	return nil
}

func (m *MemStorage) KeyExist(mType, key string) bool {
	switch mType {
	case gaugeType:
		_, ok := m.Gauge[key]
		if ok {
			return true
		}
	case counterType:
		_, ok := m.Counter[key]
		if ok {
			return true
		}
	}

	return false
}

func (m *MemStorage) GetGauge() map[string]float64 {
	return m.Gauge
}

func (m *MemStorage) GetGaugeValue(key string) float64 {
	return m.Gauge[key]
}

func (m *MemStorage) GetCounters() map[string]int64 {
	return m.Counter
}

func (m *MemStorage) GetCounterValue(key string) int64 {
	return m.Counter[key]
}

func NewMemStorage() *MemStorage {
	var m MemStorage
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)

	return &m
}
