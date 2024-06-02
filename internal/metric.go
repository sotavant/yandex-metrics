package internal

// Названия типа метрик, которыми оперирует приложение
const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

// Metrics структура для хранения метрик
type Metrics struct {
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}
