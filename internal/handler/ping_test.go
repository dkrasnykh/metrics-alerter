package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	
	mock_service "github.com/dkrasnykh/metrics-alerter/internal/service/mocks"
)

func TestHandleGetPing(t *testing.T) {
	type mockBehavior func(r *mock_service.MockServicer)

	c := gomock.NewController(t)
	defer c.Finish()

	s := mock_service.NewMockServicer(c)
	h := New(s, "", nil)

	testServ := httptest.NewServer(h.InitRoutes())
	defer testServ.Close()

	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name: "200 OK",
			mockBehavior: func(r *mock_service.MockServicer) {
				r.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "500 database connection error",
			mockBehavior: func(r *mock_service.MockServicer) {
				r.EXPECT().Ping(gomock.Any()).Return(errors.New(`internal error`))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(s)

			r := chi.NewMux()
			r.Get("/ping", h.HandleGetPing)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/ping", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.expectedStatusCode)
		})
	}
}
