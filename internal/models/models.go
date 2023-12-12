package models

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type Gauge struct {
	Name  string
	Value float64
}

type Counter struct {
	Name  string
	Value int64
}
