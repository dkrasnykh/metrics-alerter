package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/go-resty/resty/v2"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

type Agent struct {
	client         *resty.Client
	serverAddress  string
	pollInterval   int
	reportInterval int
	pollTicker     *time.Ticker
	reportTicker   *time.Ticker
	pollCount      int64
	memStats       *runtime.MemStats
}

func New(serverAddress string, pollInterval, reportInterval int) *Agent {
	return &Agent{
		client:         resty.New(),
		serverAddress:  serverAddress,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		memStats:       &runtime.MemStats{},
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
		runtime.ReadMemStats(a.memStats)
		log.Printf("metrics collection, timestamp: %s\n", t.String())
	}
}

func (a *Agent) reportMemStats() {
	for t := range a.reportTicker.C {
		log.Printf("metrics reporting, timestamp: %s", t.String())
		f := rand.Float64()
		metrics := []models.Metrics{
			{ID: `PollCount`, MType: models.CounterType, Delta: &a.pollCount},
			{ID: `RandomValue`, MType: models.GaugeType, Value: &f},
			{ID: `Alloc`, MType: models.GaugeType, Value: a.parse(a.memStats.Alloc)},
			{ID: `BuckHashSys`, MType: models.GaugeType, Value: a.parse(a.memStats.BuckHashSys)},
			{ID: `Frees`, MType: models.GaugeType, Value: a.parse(a.memStats.Frees)},
			{ID: `GCCPUFraction`, MType: models.GaugeType, Value: &a.memStats.GCCPUFraction},
			{ID: `GCSys`, MType: models.GaugeType, Value: a.parse(a.memStats.GCSys)},
			{ID: `HeapAlloc`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapAlloc)},
			{ID: `HeapIdle`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapIdle)},
			{ID: `HeapInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapInuse)},
			{ID: `HeapObjects`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapObjects)},
			{ID: `HeapReleased`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapReleased)},
			{ID: `HeapSys`, MType: models.GaugeType, Value: a.parse(a.memStats.HeapSys)},
			{ID: `LastGC`, MType: models.GaugeType, Value: a.parse(a.memStats.LastGC)},
			{ID: `Lookups`, MType: models.GaugeType, Value: a.parse(a.memStats.Lookups)},
			{ID: `MCacheInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.MCacheInuse)},
			{ID: `MCacheSys`, MType: models.GaugeType, Value: a.parse(a.memStats.MCacheSys)},
			{ID: `MSpanInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.MSpanInuse)},
			{ID: `MSpanSys`, MType: models.GaugeType, Value: a.parse(a.memStats.MSpanSys)},
			{ID: `Mallocs`, MType: models.GaugeType, Value: a.parse(a.memStats.Mallocs)},
			{ID: `NextGC`, MType: models.GaugeType, Value: a.parse(a.memStats.NextGC)},
			{ID: `NumForcedGC`, MType: models.GaugeType, Value: a.parse(uint64(a.memStats.NumForcedGC))},
			{ID: `NumGC`, MType: models.GaugeType, Value: a.parse(uint64(a.memStats.NumGC))},
			{ID: `OtherSys`, MType: models.GaugeType, Value: a.parse(a.memStats.OtherSys)},
			{ID: `PauseTotalNs`, MType: models.GaugeType, Value: a.parse(a.memStats.PauseTotalNs)},
			{ID: `StackInuse`, MType: models.GaugeType, Value: a.parse(a.memStats.StackInuse)},
			{ID: `StackSys`, MType: models.GaugeType, Value: a.parse(a.memStats.StackSys)},
			{ID: `Sys`, MType: models.GaugeType, Value: a.parse(a.memStats.Sys)},
			{ID: `TotalAlloc`, MType: models.GaugeType, Value: a.parse(a.memStats.TotalAlloc)},
		}
		for _, m := range metrics {
			go a.sendRequest(m)
		}
	}
}

func (a *Agent) parse(v uint64) *float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	if err != nil {
		log.Printf("error: %s", err.Error())
	}
	return &f
}

func (a *Agent) sendRequest(m models.Metrics) {
	checkAndLog := func(err error) {
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
	}
	url := fmt.Sprintf("http://%s/update/", a.serverAddress)
	requestBody, err := json.Marshal(m)
	checkAndLog(err)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	defer func(gz *gzip.Writer) {
		err := gz.Close()
		checkAndLog(err)
	}(gz)
	_, err = gz.Write(requestBody)
	checkAndLog(err)
	err = gz.Close()
	checkAndLog(err)
	buf := b.Bytes()

	resp, err := a.client.R().SetHeader(headers.ContentType, `application/json`).
		SetHeader(headers.ContentEncoding, `gzip`).
		SetHeader(headers.AcceptEncoding, `gzip`).
		SetBody(buf).Post(url)
	checkAndLog(err)
	if resp.StatusCode() != http.StatusOK {
		log.Printf("error: %s", err.Error())
	}
}
