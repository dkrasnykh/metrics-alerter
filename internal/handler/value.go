package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

func (h *Handler) HandleGetByParam(res http.ResponseWriter, req *http.Request) {
	metricType, metricName := chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName")
	res.Header().Set(headers.ContentType, "text/plain")

	value, err := h.service.GetMetricValue(req.Context(), metricType, metricName)

	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = res.Write([]byte(value))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGet(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(headers.ContentType, "application/json")
	m, err := extractBody(req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if m.MType != models.CounterType && m.MType != models.GaugeType {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	*m, err = h.service.Get(req.Context(), (*m).MType, (*m).ID)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	buf, err := json.Marshal(*m)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = res.Write(buf)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func extractBody(req *http.Request) (*models.Metrics, error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 || string(bytes) == `null` {
		return nil, errors.New(`request body is empty`)
	}
	var m models.Metrics
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
