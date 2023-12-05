package handler

import (
	"fmt"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUpdate(t *testing.T) {
	r := storage.NewStorage()
	v := service.NewService(r)
	h := NewHandler(v)
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
