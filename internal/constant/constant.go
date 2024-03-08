package constant

const(
	CounterS     = "counter"
	GaugeS       = "gauge"
)
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type Gauge float64
type Counter int64

type GaugeMetric map[string]Gauge
type CounterMetric map[string]Counter
