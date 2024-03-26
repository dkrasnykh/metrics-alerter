package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
)

func (h *Handler) HandleUpdateByParam(res http.ResponseWriter, req *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(req, "metricType"),
		chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")

	res.Header().Set(headers.ContentType, "text/plain")

	m := convert(metricType, metricName, metricValue)
	err := h.service.Validate(m)
	if err != nil {
		if errors.Is(err, service.ErrIDIsEmpty) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.service.Save(req.Context(), m)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleUpdate(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(headers.ContentType, "application/json")
	m, err := extractBody(req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.service.Validate(*m)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	*m, err = h.service.Save(req.Context(), *m)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
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

func convert(mtype, mname, value string) models.Metrics {
	m := models.Metrics{MType: mtype, ID: mname}
	switch mtype {
	case models.CounterType:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			m.Delta = &delta
		}
	case models.GaugeType:
		gvalue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			m.Value = &gvalue
		}
	}
	return m
}
