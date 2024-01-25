package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

func TestHandleUpdateByParam(t *testing.T) {
	_ = logger.InitLogger()
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``)
	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	tests := []struct {
		name        string
		request     string
		code        int
		contentType string
	}{
		{
			name:        "success update counter",
			request:     "/update/counter/test/100",
			code:        http.StatusOK,
			contentType: "text/plain",
		},
		{
			name:        "success update gauge",
			request:     "/update/gauge/test/100",
			code:        http.StatusOK,
			contentType: "text/plain",
		},
		{
			name:        "invalid url - unidentified metric type",
			request:     "/update/test/test/100",
			code:        http.StatusBadRequest,
			contentType: "text/plain",
		},
		{
			name:        "invalid url - bad value",
			request:     "/update/test/test/test",
			code:        http.StatusBadRequest,
			contentType: "text/plain",
		},
		{
			name:        "invalid url - metric name is empty",
			request:     "/update/gauge//100",
			code:        http.StatusNotFound,
			contentType: "text/plain",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			url := fmt.Sprintf("%s%s", testServ.URL, test.request)
			req, err := http.NewRequest(http.MethodPost, url, nil)
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
			assert.Empty(t, string(respBody))
		})
	}
}

func TestHandleGetByParam(t *testing.T) {
	_ = logger.InitLogger()
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``)
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
