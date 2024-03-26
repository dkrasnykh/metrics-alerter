package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	mock_service "github.com/dkrasnykh/metrics-alerter/internal/service/mocks"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

func TestHandleUpdateByParam(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``, nil)
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

func TestHandleUpdate(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``, nil)
	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()
	URL := fmt.Sprintf("%s%s", testServ.URL, "/update/")

	delta := int64(123)
	value := float64(123)
	mCounter := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	mGauge := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}

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
			name:        "400 bad request - indefined metric type",
			requestBody: models.Metrics{MType: "indefined", ID: "id"},
			code:        http.StatusBadRequest,
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

func TestHandleUpdateInternalError(t *testing.T) {
	type mockBehavior func(r *mock_service.MockServicer, metric models.Metrics)

	c := gomock.NewController(t)
	defer c.Finish()

	s := mock_service.NewMockServicer(c)
	h := New(s, "", nil)

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	delta := int64(123)
	m := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}

	tests := []struct {
		name               string
		requestBody        models.Metrics
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name:        "500 internal server error",
			requestBody: m,
			mockBehavior: func(r *mock_service.MockServicer, metric models.Metrics) {
				r.EXPECT().Validate(metric).Return(nil)
				r.EXPECT().Save(gomock.Any(), metric).Return(models.Metrics{}, errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(s, test.requestBody)

			r := chi.NewMux()
			r.Post("/update/", h.HandleUpdate)

			mExpected, err := json.Marshal(test.requestBody)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(mExpected))

			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.expectedStatusCode)
		})
	}
}

func TestConvert(t *testing.T) {
	delta := int64(123)
	mCounterExpected := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	mCounterActual := convert(models.CounterType, "testCounter", "123")
	assert.Equal(t, mCounterExpected, mCounterActual)

	value := float64(123)
	mGaugeExpected := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}
	mGaugeActual := convert(models.GaugeType, "testGuade", "123")
	assert.Equal(t, mGaugeExpected, mGaugeActual)
}
