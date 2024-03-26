package handler

import (
	"context"
	"errors"
	"fmt"
	"html/template"
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
	"github.com/dkrasnykh/metrics-alerter/internal/service/mocks"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

func TestHandleGetAll(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	tpl, err := template.New("webpage").Parse(Tpl)
	require.NoError(t, err)
	h := New(v, ``, tpl)
	ctx := context.Background()

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	delta := int64(123)
	value := float64(123)
	m1 := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	m2 := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}

	_, err = r.Create(ctx, m1)
	require.NoError(t, err)
	_, err = r.Create(ctx, m2)
	require.NoError(t, err)

	tests := []struct {
		name     string
		request  string
		code     int
		response string
	}{
		{
			name:     `200 OK`,
			request:  "/",
			code:     http.StatusOK,
			response: "\n\t<!DOCTYPE html>\n\t<html>\n\t\t<body>\n\t\t\t<div>counter testCounter 123 &lt;nil&gt;</div><div>gauge testGuade &lt;nil&gt; 123</div>\n\t\t</body>\n\t</html>",
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

func TestHandleGetAllExecuteTemplateError(t *testing.T) {
	r := memory.New("", 0)
	v := service.New(r)
	tpl := template.New("webpage")

	h := New(v, ``, tpl)
	ctx := context.Background()

	delta := int64(123)
	value := float64(123)
	m1 := models.Metrics{MType: models.CounterType, ID: `testCounter`, Delta: &delta}
	m2 := models.Metrics{MType: models.GaugeType, ID: `testGuade`, Value: &value}

	_, err := r.Create(ctx, m1)
	require.NoError(t, err)
	_, err = r.Create(ctx, m2)
	require.NoError(t, err)

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	tests := []struct {
		name    string
		request string
		code    int
	}{
		{
			name:    "execute html template error",
			request: "/",
			code:    http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", testServ.URL, test.request)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := testServ.Client().Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.code, resp.StatusCode)
		})
	}
}

func TestHandleGetAllInternalServiceError(t *testing.T) {
	type mockBehavior func(r *mock_service.MockServicer)

	c := gomock.NewController(t)
	defer c.Finish()

	s := mock_service.NewMockServicer(c)
	h := New(s, "", nil)

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	tests := []struct {
		name               string
		request            string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name:    "ok",
			request: "/",
			mockBehavior: func(r *mock_service.MockServicer) {
				r.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(s)

			r := chi.NewMux()
			r.Get("/", h.HandleGetAll)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.expectedStatusCode)
		})
	}
}
