package models

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
	Attempts    uint   = 3
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func DelayType(n uint, _ error, config *retry.Config) time.Duration {
	switch n {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	default:
		return 5 * time.Second
	}
}

func OnRetry(n uint, err error) {
	logger.Error(fmt.Sprintf(`%d %w`, n, err))
}
