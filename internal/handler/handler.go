package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
)

const (
	Tpl = `
	<!DOCTYPE html>
	<html>
		<body>
			{{range .Metrics}}<div>{{ .MType }} {{ .ID }} {{ .Delta }} {{ .Value }}</div>{{end}}
		</body>
	</html>`
)

var T *template.Template

type Handler struct {
	service *service.Service
	key     string
}

func New(s *service.Service, key string) *Handler {
	return &Handler{
		service: s,
		key:     key,
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(h.Hash)
	r.Use(h.GzipRequest)
	r.Use(h.GzipResponse)
	r.Use(h.Logging)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdateByParam)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetByParam)
	r.Get("/", h.HandleGetAll)
	r.Post("/update/", h.HandleUpdate)
	r.Post("/value/", h.HandleGet)
	r.Get("/ping", h.HandleGetPing)
	r.Post("/updates/", h.HandleUpdates)

	return r
}

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

func (h *Handler) HandleGetAll(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(headers.ContentType, `text/html`)
	type Item struct {
		Metrics []models.Metrics
	}
	metrics, err := h.service.GetAll(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = T.Execute(res, Item{Metrics: metrics})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
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

func (h *Handler) HandleGetPing(res http.ResponseWriter, req *http.Request) {
	err := h.service.Ping(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) HandleUpdates(res http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metrics := []models.Metrics{}
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, m := range metrics {
		err = h.service.Validate(m)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	err = h.service.Load(req.Context(), metrics)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func extractBody(req *http.Request) (*models.Metrics, error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, errors.New(`request body is empty`)
	}
	var m models.Metrics
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
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
