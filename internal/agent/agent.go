package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
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
	pollCount      int
	memStats       *runtime.MemStats
}

func NewAgent(serverAddress string, pollInterval, reportInterval int) *Agent {
	return &Agent{
		client:         resty.New(),
		serverAddress:  serverAddress,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		memStats:       &runtime.MemStats{},
	}
}

func (a *Agent) Run() error {
	a.pollTicker = time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer a.pollTicker.Stop()
	a.reportTicker = time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer a.reportTicker.Stop()

	go a.collectMemStats()

	ch := make(chan error)
	go a.reportMemStats(ch)

	err := <-ch
	return err
}

func (a *Agent) collectMemStats() {
	for t := range a.pollTicker.C {
		a.pollCount++
		runtime.ReadMemStats(a.memStats)
		log.Printf("metrics collection, timestamp: %s\n", t.String())
	}
}

func (a *Agent) reportMemStats(ch chan error) {
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
				Type: models.CounterType,
				Values: []Value{
					{"pollCount", fmt.Sprintf("%d", a.pollCount)},
				},
			},
			{
				Type: models.GaugeType,
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
				go a.sendRequest(mType, m.Name, m.Value, ch)
			}
		}
	}
}

func (a *Agent) sendRequest(metricType, metricName, metricValue string, ch chan error) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.serverAddress, metricType, metricName, metricValue)
	resp, err := a.client.R().SetHeader(headers.ContentType, "text/plain").Post(url)
	if err != nil {
		ch <- fmt.Errorf("failed to handle response from server:  %s\n%s", url, err.Error())
	}
	if resp.StatusCode() != http.StatusOK {
		ch <- fmt.Errorf("unexpected response status %d from request %s", resp.StatusCode(), url)
	}
}
