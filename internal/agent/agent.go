package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-http-utils/headers"
	"github.com/go-resty/resty/v2"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/utils"
)

type SyncMemStats struct {
	v  *runtime.MemStats
	mx sync.RWMutex
}

type Agent struct {
	client         *resty.Client
	serverAddress  string
	pollInterval   int
	reportInterval int
	pollTicker     *time.Ticker
	reportTicker   *time.Ticker
	pollCount      int64
	memStats       SyncMemStats
}

func New(serverAddress string, pollInterval, reportInterval int) *Agent {
	return &Agent{
		client:         resty.New(),
		serverAddress:  serverAddress,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		memStats: SyncMemStats{
			v:  &runtime.MemStats{},
			mx: sync.RWMutex{},
		},
	}
}

func (a *Agent) Run() {
	a.pollTicker = time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer a.pollTicker.Stop()
	a.reportTicker = time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer a.reportTicker.Stop()

	go a.collectMemStats()
	go a.reportMemStats()

	time.Sleep(time.Minute)
}

func (a *Agent) collectMemStats() {
	for t := range a.pollTicker.C {
		a.pollCount++

		a.memStats.mx.Lock()
		runtime.ReadMemStats(a.memStats.v)
		a.memStats.mx.Unlock()

		logger.Info(fmt.Sprintf("metrics collection, timestamp: %s\n", t.String()))
	}
}

func (a *Agent) reportMemStats() {
	for t := range a.reportTicker.C {
		logger.Info(fmt.Sprintf("metrics reporting, timestamp: %s", t.String()))
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
		}
		a.memStats.mx.RUnlock()

		go a.sendBatchRequest(metrics)
	}
}

func (a *Agent) parse(v uint64) *float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	utils.LogError(err)
	return &f
}

func (a *Agent) sendBatchRequest(metrics []models.Metrics) {
	buf := gzipData(metrics)
	var resp *resty.Response
	err := retry.Do(
		func() error {
			var err error
			resp, err = a.client.R().SetHeader(headers.ContentType, `application/json`).
				SetHeader(headers.ContentEncoding, `gzip`).
				SetHeader(headers.AcceptEncoding, `gzip`).
				SetBody(buf).Post(fmt.Sprintf("http://%s/updates/", a.serverAddress))
			return err
		},
		retry.Attempts(config.Attempts),
		retry.DelayType(config.DelayType),
		retry.OnRetry(config.OnRetry),
	)
	utils.LogError(err)
	if resp.StatusCode() != http.StatusOK {
		logger.Error(fmt.Sprintf(`unexpected status code %d`, resp.StatusCode()))
	}
}

func gzipData(any interface{}) []byte {
	body, err := json.Marshal(any)
	utils.LogError(err)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write(body)
	utils.LogError(err)
	err = gz.Close()
	utils.LogError(err)
	buf := b.Bytes()
	return buf
}
