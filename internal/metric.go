package internal

// Названия типа метрик, которыми оперирует приложение
const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

// Metrics структура для хранения метрик
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
