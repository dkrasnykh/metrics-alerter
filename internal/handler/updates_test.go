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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	mock_service "github.com/dkrasnykh/metrics-alerter/internal/service/mocks"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

func TestHandleUpdates(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	h := New(v, ``, nil)
	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()
	URL := fmt.Sprintf("%s%s", testServ.URL, "/updates/")

	delta := int64(123)
	value := float64(123)
	mCounter := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	mGauge := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}

	metrics := []models.Metrics{mCounter, mGauge}

	tests := []struct {
		name        string
		requestBody []models.Metrics
		code        int
	}{
		{
			name:        "200 OK",
			requestBody: metrics,
			code:        http.StatusOK,
		},
		{
			name:        "400 bad request - invalid request body",
			requestBody: nil,
			code:        http.StatusBadRequest,
		},
		{
			name:        "400 bad request - validation failure",
			requestBody: []models.Metrics{{MType: "invalid", ID: "2"}},
			code:        http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsExpected, err := json.Marshal(test.requestBody)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(metricsExpected))
			require.NoError(t, err)

			resp, err := testServ.Client().Do(req)
			require.NoError(t, err)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, test.code, resp.StatusCode)
			assert.Empty(t, respBody)
		})
	}
}

func TestHandleUpdatesMock(t *testing.T) {
	type mockBehavior func(r *mock_service.MockServicer, metrics []models.Metrics)

	c := gomock.NewController(t)
	defer c.Finish()

	s := mock_service.NewMockServicer(c)
	h := New(s, "", nil)

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	delta := int64(123)
	value := float64(123)
	mCounter := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	mGauge := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}

	metrics := []models.Metrics{mCounter, mGauge}

	tests := []struct {
		name         string
		requestBody  []models.Metrics
		mockBehavior mockBehavior
		code         int
	}{
		{
			name:        "500 internal server error - load failure",
			requestBody: metrics,
			mockBehavior: func(r *mock_service.MockServicer, metrics []models.Metrics) {
				r.EXPECT().Validate(metrics[0]).Return(nil)
				r.EXPECT().Validate(metrics[1]).Return(nil)
				r.EXPECT().Load(gomock.Any(), metrics).Return(errors.New("internal error"))
			},
			code: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(s, test.requestBody)

			r := chi.NewMux()
			r.Post("/updates/", h.HandleUpdates)

			metricsJson, err := json.Marshal(test.requestBody)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/updates/", bytes.NewBuffer(metricsJson))

			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.code)
		})
	}

}
