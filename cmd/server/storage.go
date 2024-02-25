package main

type Storage interface {
	AddGaugeValue(key string, value float64)
	AddCounterValue(key string, value int64)
	GetValue(mType string, key string) interface{}
	GetGauge() map[string]float64
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

func (m *MemStorage) GetGauge() map[string]float64 {
	return m.Gauge
}

func NewMemStorage() *MemStorage {
	var m MemStorage
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)

	return &m
}
