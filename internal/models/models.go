package models

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type Metric struct {
	Type         string
	Name         string
	ValueInt64   int64
	ValueFloat64 float64
}
