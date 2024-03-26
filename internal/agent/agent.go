package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-http-utils/headers"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/hash"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type SyncMemStats struct {
	v               *runtime.MemStats
	TotalMemory     float64
	FreeMemory      float64
	CPUutilization1 float64
	mx              sync.RWMutex
}

type Agent struct {
	client         *resty.Client
	serverAddress  string
	pollInterval   int
	reportInterval int
	pollTicker     *time.Ticker
	reportTicker   *time.Ticker
	pollCount      int64
	key            string
	rateLimit      int
	memStats       SyncMemStats
}

func New(c *config.AgentConfig) *Agent {
	return &Agent{
		client:         resty.New(),
		serverAddress:  c.Address,
		pollInterval:   c.PollInterval,
		reportInterval: c.ReportInterval,
		key:            c.Key,
		rateLimit:      c.RateLimit,
		memStats: SyncMemStats{
			v:  &runtime.MemStats{},
			mx: sync.RWMutex{},
		},
	}
}

func (a *Agent) Run(ctx context.Context) {
	a.pollTicker = time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer a.pollTicker.Stop()
	a.reportTicker = time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer a.reportTicker.Stop()

	go a.collectMemStats()

	jobs := make(chan []models.Metrics)
	results := make(chan struct{})
	for w := 1; w <= a.rateLimit; w++ {
		go a.worker(ctx, jobs, results)
	}

	go a.reportMemStats(jobs)

	<-results
}

func (a *Agent) collectMemStats() {
	for t := range a.pollTicker.C {
		a.pollCount++

		a.memStats.mx.Lock()

		runtime.ReadMemStats(a.memStats.v)
		vm, err := mem.VirtualMemory()
		if err != nil {
			zap.L().Error(err.Error())
		}
		a.memStats.FreeMemory = float64(vm.Free)
		a.memStats.TotalMemory = float64(vm.Total)
		cp, err := cpu.Percent(time.Millisecond, false)
		if err != nil {
			zap.L().Error(err.Error())
		}
		a.memStats.CPUutilization1 = cp[0]

		a.memStats.mx.Unlock()

		zap.L().Info(fmt.Sprintf("metrics collection, timestamp: %s\n", t.String()))
	}
}

func (a *Agent) worker(ctx context.Context, jobs <-chan []models.Metrics, results chan struct{}) {
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				results <- struct{}{}
			}
			a.sendBatchRequest(job)
		case <-ctx.Done():
			results <- struct{}{}
		}
	}
}

func (a *Agent) reportMemStats(jobs chan []models.Metrics) {
	for t := range a.reportTicker.C {
		zap.L().Info(fmt.Sprintf("metrics reporting, timestamp: %s", t.String()))
		f := rand.Float64()

		a.memStats.mx.RLock()
		metrics := []models.Metrics{
			{ID: `PollCount`, MType: models.CounterType, Delta: &a.pollCount},
			{ID: `RandomValue`, MType: models.GaugeType, Value: &f},
			{ID: `Alloc`, MType: models.GaugeType, Value: a.parse(a.memStats.v.Alloc)},
			{ID: `BuckHashSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.BuckHashSys)},
			{ID: `Frees`, MType: models.GaugeType, Value: a.parse(a.memStats.v.Frees)},
			{ID: `GCCPUFraction`, MType: models.GaugeType, Value: &a.memStats.v.GCCPUFraction},
			{ID: `GCSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.GCSys)},
			{ID: `HeapAlloc`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapAlloc)},
			{ID: `HeapIdle`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapIdle)},
			{ID: `HeapInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapInuse)},
			{ID: `HeapObjects`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapObjects)},
			{ID: `HeapReleased`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapReleased)},
			{ID: `HeapSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.HeapSys)},
			{ID: `LastGC`, MType: models.GaugeType, Value: a.parse(a.memStats.v.LastGC)},
			{ID: `Lookups`, MType: models.GaugeType, Value: a.parse(a.memStats.v.Lookups)},
			{ID: `MCacheInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.v.MCacheInuse)},
			{ID: `MCacheSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.MCacheSys)},
			{ID: `MSpanInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.v.MSpanInuse)},
			{ID: `MSpanSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.MSpanSys)},
			{ID: `Mallocs`, MType: models.GaugeType, Value: a.parse(a.memStats.v.Mallocs)},
			{ID: `NextGC`, MType: models.GaugeType, Value: a.parse(a.memStats.v.NextGC)},
			{ID: `NumForcedGC`, MType: models.GaugeType, Value: a.parse(uint64(a.memStats.v.NumForcedGC))},
			{ID: `NumGC`, MType: models.GaugeType, Value: a.parse(uint64(a.memStats.v.NumGC))},
			{ID: `OtherSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.OtherSys)},
			{ID: `PauseTotalNs`, MType: models.GaugeType, Value: a.parse(a.memStats.v.PauseTotalNs)},
			{ID: `StackInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.v.StackInuse)},
			{ID: `StackSys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.StackSys)},
			{ID: `Sys`, MType: models.GaugeType, Value: a.parse(a.memStats.v.Sys)},
			{ID: `TotalAlloc`, MType: models.GaugeType, Value: a.parse(a.memStats.v.TotalAlloc)},
			{ID: `TotalMemory`, MType: models.GaugeType, Value: &a.memStats.TotalMemory},
			{ID: `FreeMemory`, MType: models.GaugeType, Value: &a.memStats.FreeMemory},
			{ID: `CPUutilization1`, MType: models.GaugeType, Value: &a.memStats.CPUutilization1},
		}
		a.memStats.mx.RUnlock()

		jobs <- metrics
	}
}

func (a *Agent) parse(v uint64) *float64 {
	f := float64(v)
	return &f
}

func (a *Agent) sendBatchRequest(metrics []models.Metrics) {
	req := a.client.R().SetHeader(headers.ContentType, `application/json`).
		SetHeader(headers.ContentEncoding, `gzip`).
		SetHeader(headers.AcceptEncoding, `gzip`)
	buf := gzipData(metrics)
	if a.key != "" {
		req.SetHeader(hash.Header, hash.Encode(buf, []byte(a.key)))
	}
	req.SetBody(buf)

	var resp *resty.Response
	err := retry.Do(
		func() error {
			var err error
			resp, err = req.Post(fmt.Sprintf("http://%s/updates/", a.serverAddress))
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	if err != nil {
		zap.L().Error(err.Error())
	}
	if resp.StatusCode() != http.StatusOK {
		zap.L().Error(fmt.Sprintf(`unexpected status code %d`, resp.StatusCode()))
	}
}

func gzipData(any interface{}) []byte {
	body, err := json.Marshal(any)
	if err != nil {
		zap.L().Error(err.Error())
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write(body)
	if err != nil {
		zap.L().Error(err.Error())
	}
	err = gz.Close()
	if err != nil {
		zap.L().Error(err.Error())
	}
	buf := b.Bytes()
	return buf
}
