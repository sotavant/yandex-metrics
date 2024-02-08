package internal

type Storage interface {
	AddGaugeValue(key string, value float64)
	AddCounterValue(key string, value int64)
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

func NewMemStorage() *MemStorage {
	var m MemStorage
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)

	return &m
}
