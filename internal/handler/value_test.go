package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

func TestHandleGetByParam(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``, nil)
	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()
	delta := int64(123)
	value := float64(123)
	m1 := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	m2 := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}
	_, err := r.Create(context.Background(), m1)
	require.NoError(t, err)
	_, err = r.Create(context.Background(), m2)
	require.NoError(t, err)

	tests := []struct {
		name     string
		request  string
		code     int
		response string
	}{
		{
			name:     "success getting counter value",
			request:  "/value/counter/testCounter",
			code:     http.StatusOK,
			response: "123",
		},
		{
			name:     "success getting gauge value",
			request:  "/value/gauge/testGuade",
			code:     http.StatusOK,
			response: "123",
		},
		{
			name:     "unknown metric name",
			request:  "/value/gauge/unknown",
			code:     http.StatusNotFound,
			response: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", testServ.URL, test.request)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := testServ.Client().Do(req)
			require.NoError(t, err)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, test.code, resp.StatusCode)
			assert.Equal(t, test.response, string(respBody))
		})
	}
}

func TestHandleGet(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``, nil)
	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()
	URL := fmt.Sprintf("%s%s", testServ.URL, "/value/")

	delta := int64(123)
	value := float64(123)
	mCounter := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	mGauge := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}
	_, err := r.Create(context.Background(), mCounter)
	require.NoError(t, err)
	_, err = r.Create(context.Background(), mGauge)
	require.NoError(t, err)

	tests := []struct {
		name        string
		requestBody models.Metrics
		code        int
		response    string
		contentType string
	}{
		{
			name:        "200 OK conter type",
			requestBody: mCounter,
			code:        http.StatusOK,
			response:    `{"id":"testCounter","type":"counter","delta":123}`,
			contentType: "application/json",
		},
		{
			name:        "200 OK gauge type",
			requestBody: mGauge,
			code:        http.StatusOK,
			response:    `{"id":"testGuade","type":"gauge","value":123}`,
			contentType: "application/json",
		},
		{
			name:        "400 bad request - invalid metric type",
			requestBody: models.Metrics{MType: "invalid", ID: "id"},
			code:        http.StatusBadRequest,
			response:    "",
			contentType: "application/json",
		},
		{
			name:        "404 not found",
			requestBody: models.Metrics{MType: models.CounterType, ID: "new"},
			code:        http.StatusNotFound,
			response:    "",
			contentType: "application/json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mExpected, err := json.Marshal(test.requestBody)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(mExpected))
			require.NoError(t, err)

			resp, err := testServ.Client().Do(req)
			require.NoError(t, err)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			contentType := resp.Header.Get(headers.ContentType)

			assert.Equal(t, test.code, resp.StatusCode)
			assert.Equal(t, test.contentType, contentType)
			assert.Equal(t, test.response, string(respBody))
		})
	}
}
