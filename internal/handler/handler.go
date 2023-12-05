package handler

import (
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{s}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdate)

	return router
}

func (h *Handler) HandleUpdate(res http.ResponseWriter, req *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(req, "metricType"),
		chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")
	res.Header().Set(headers.ContentType, "text/plain")
	if metricName == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	err := h.service.ValidateAndSave(metricType, metricName, metricValue)

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.WriteHeader(http.StatusOK)
	}
}
