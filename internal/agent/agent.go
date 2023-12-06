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
	memStats      *runtime.MemStats
}

func NewAgent(serverAddress, serverPort string, pollTicker, reportTicker *time.Ticker) *Agent {
	return &Agent{
		client:        resty.New(),
		serverAddress: serverAddress,
		serverPort:    serverPort,
		pollTicker:    pollTicker,
		reportTicker:  reportTicker,
		memStats:      &runtime.MemStats{},
	}
}

func (a *Agent) Run() {
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
	type Value struct {
		Name  string
		Value string
	}

	type Item struct {
		Type   string
		Values []Value
	}

	for t := range a.reportTicker.C {
		log.Printf("metrics reporting, timestamp: %s", t.String())
		items := []Item{
			{
				Type: storage.Counter,
				Values: []Value{
					{"pollCount", fmt.Sprintf("%d", a.pollCount)},
				},
			},
			{
				Type: storage.Gauge,
				Values: []Value{
					{"RandomValue", fmt.Sprintf("%d", rand.Intn(10000))},
					{"Alloc", fmt.Sprintf("%d", a.memStats.Alloc)},
					{"BuckHashSys", fmt.Sprintf("%d", a.memStats.BuckHashSys)},
					{"Frees", fmt.Sprintf("%d", a.memStats.Frees)},
					{"GCCPUFraction", fmt.Sprintf("%f", a.memStats.GCCPUFraction)},
					{"GCSys", fmt.Sprintf("%d", a.memStats.GCSys)},
					{"HeapAlloc", fmt.Sprintf("%d", a.memStats.HeapAlloc)},
					{"HeapIdle", fmt.Sprintf("%d", a.memStats.HeapIdle)},
					{"HeapInuse", fmt.Sprintf("%d", a.memStats.HeapInuse)},
					{"HeapObjects", fmt.Sprintf("%d", a.memStats.HeapObjects)},
					{"HeapReleased", fmt.Sprintf("%d", a.memStats.HeapReleased)},
					{"HeapSys", fmt.Sprintf("%d", a.memStats.HeapSys)},
					{"LastGC", fmt.Sprintf("%d", a.memStats.LastGC)},
					{"Lookups", fmt.Sprintf("%d", a.memStats.Lookups)},
					{"MCacheInuse", fmt.Sprintf("%d", a.memStats.MCacheInuse)},
					{"MCacheSys", fmt.Sprintf("%d", a.memStats.MCacheSys)},
					{"MSpanInuse", fmt.Sprintf("%d", a.memStats.MSpanInuse)},
					{"Mallocs", fmt.Sprintf("%d", a.memStats.Mallocs)},
					{"NextGC", fmt.Sprintf("%d", a.memStats.NextGC)},
					{"NumForcedGC", fmt.Sprintf("%d", a.memStats.NumForcedGC)},
					{"NumGC", fmt.Sprintf("%d", a.memStats.NumGC)},
					{"OtherSys", fmt.Sprintf("%d", a.memStats.OtherSys)},
					{"PauseTotalNs", fmt.Sprintf("%d", a.memStats.PauseTotalNs)},
					{"StackInuse", fmt.Sprintf("%d", a.memStats.StackInuse)},
					{"StackSys", fmt.Sprintf("%d", a.memStats.StackSys)},
					{"Sys", fmt.Sprintf("%d", a.memStats.Sys)},
					{"TotalAlloc", fmt.Sprintf("%d", a.memStats.TotalAlloc)},
				},
			},
		}
		for _, item := range items {
			mType := item.Type
			for _, m := range item.Values {
				go a.sendRequest(mType, m.Name, m.Value)
			}
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
