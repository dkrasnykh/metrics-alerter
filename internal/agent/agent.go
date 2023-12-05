package agent

import (
	"fmt"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"github.com/go-http-utils/headers"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Agent struct {
	client        *resty.Client
	serverAddress string
	serverPort    string
	pollTicker    *time.Ticker
	reportTicker  *time.Ticker
	pollCount     int
	metrics       map[storage.KeyStorage]string
}

func NewAgent(serverAddress, serverPort string, pollTicker, reportTicker *time.Ticker) *Agent {
	return &Agent{
		client:        resty.New(),
		serverAddress: serverAddress,
		serverPort:    serverPort,
		pollTicker:    pollTicker,
		reportTicker:  reportTicker,
		metrics:       make(map[storage.KeyStorage]string),
	}
}

func (a *Agent) Run() {
	go a.collectMemStats()
	go a.reportMemStats()
	time.Sleep(time.Minute)
}

func (a *Agent) reportMemStats() {
	for t := range a.reportTicker.C {
		log.Printf("metrics reporting, timestamp: %s", t.String())
		for k, v := range a.metrics {
			go a.sendRequest(k.MetricType, k.MetricName, v)
		}
	}
}

func (a *Agent) sendRequest(metricType, metricName, metricValue string) {
	url := fmt.Sprintf("http://%s:%s/update/%s/%s/%s", a.serverAddress, a.serverPort, metricType, metricName, metricValue)
	resp, err := a.client.R().SetHeader(headers.ContentType, "text/plain").Post(url)
	if err != nil {
		log.Fatalf("failed to handle response from server:  %s\n%s", url, err.Error())
	}
	if resp.StatusCode() != http.StatusOK {
		log.Fatalf("unexpected response status %d from request %s", resp.StatusCode(), url)
	}
}

func (a *Agent) collectMemStats() {
	for t := range a.pollTicker.C {
		a.pollCount++

		a.metrics[storage.KeyStorage{storage.Counter, "PollCount"}] = fmt.Sprintf("%d", a.pollCount)

		randomValue := rand.Intn(10000)
		a.metrics[storage.KeyStorage{storage.Gauge, "RandomValue"}] = fmt.Sprintf("%d", randomValue)

		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		a.metrics[storage.KeyStorage{storage.Gauge, "Alloc"}] = fmt.Sprintf("%d", memStats.Alloc)
		a.metrics[storage.KeyStorage{storage.Gauge, "BuckHashSys"}] = fmt.Sprintf("%d", memStats.BuckHashSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "Frees"}] = fmt.Sprintf("%d", memStats.Frees)
		a.metrics[storage.KeyStorage{storage.Gauge, "GCCPUFraction"}] = fmt.Sprintf("%f", memStats.GCCPUFraction)
		a.metrics[storage.KeyStorage{storage.Gauge, "GCSys"}] = fmt.Sprintf("%d", memStats.GCSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapAlloc"}] = fmt.Sprintf("%d", memStats.HeapAlloc)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapIdle"}] = fmt.Sprintf("%d", memStats.HeapIdle)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapInuse"}] = fmt.Sprintf("%d", memStats.HeapInuse)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapObjects"}] = fmt.Sprintf("%d", memStats.HeapObjects)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapReleased"}] = fmt.Sprintf("%d", memStats.HeapReleased)
		a.metrics[storage.KeyStorage{storage.Gauge, "HeapSys"}] = fmt.Sprintf("%d", memStats.HeapSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "LastGC"}] = fmt.Sprintf("%d", memStats.LastGC)
		a.metrics[storage.KeyStorage{storage.Gauge, "Lookups"}] = fmt.Sprintf("%d", memStats.Lookups)
		a.metrics[storage.KeyStorage{storage.Gauge, "MCacheInuse"}] = fmt.Sprintf("%d", memStats.MCacheInuse)
		a.metrics[storage.KeyStorage{storage.Gauge, "MCacheSys"}] = fmt.Sprintf("%d", memStats.MCacheSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "MSpanInuse"}] = fmt.Sprintf("%d", memStats.MSpanInuse)
		a.metrics[storage.KeyStorage{storage.Gauge, "MSpanSys"}] = fmt.Sprintf("%d", memStats.MSpanSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "Mallocs"}] = fmt.Sprintf("%d", memStats.Mallocs)
		a.metrics[storage.KeyStorage{storage.Gauge, "NextGC"}] = fmt.Sprintf("%d", memStats.NextGC)
		a.metrics[storage.KeyStorage{storage.Gauge, "NumForcedGC"}] = fmt.Sprintf("%d", memStats.NumForcedGC)
		a.metrics[storage.KeyStorage{storage.Gauge, "NumGC"}] = fmt.Sprintf("%d", memStats.NumGC)
		a.metrics[storage.KeyStorage{storage.Gauge, "OtherSys"}] = fmt.Sprintf("%d", memStats.OtherSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "PauseTotalNs"}] = fmt.Sprintf("%d", memStats.PauseTotalNs)
		a.metrics[storage.KeyStorage{storage.Gauge, "StackInuse"}] = fmt.Sprintf("%d", memStats.StackInuse)
		a.metrics[storage.KeyStorage{storage.Gauge, "StackSys"}] = fmt.Sprintf("%d", memStats.StackSys)
		a.metrics[storage.KeyStorage{storage.Gauge, "Sys"}] = fmt.Sprintf("%d", memStats.Sys)
		a.metrics[storage.KeyStorage{storage.Gauge, "TotalAlloc"}] = fmt.Sprintf("%d", memStats.TotalAlloc)

		log.Printf("metrics collection, timestamp: %s\n", t.String())
	}
}
