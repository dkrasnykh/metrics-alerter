package handler

import (
	"errors"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"net/http"
	"strings"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{s}
}

func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	metricType, metricName, metricValue, err := parseUrl(req.URL.Path)

	if err != nil || strings.TrimSpace(metricName) == `` {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	err = h.service.ValidateAndSave(metricType, metricName, metricValue)

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.WriteHeader(http.StatusOK)
	}
}

func parseUrl(path string) (string, string, string, error) {
	parts := strings.Split(path[1:], `/`)

	if len(parts) != 4 {
		return ``, ``, ``, errors.New("invalid url")
	}

	return parts[1], parts[2], parts[3], nil
}
