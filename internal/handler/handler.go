package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
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
	logger  *logger.Logger
}

func New(s *service.Service, l *logger.Logger) *Handler {
	return &Handler{service: s,
		logger: l}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(h.GzipRequest)
	r.Use(h.GzipResponse)
	r.Use(h.Logging)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdateByParam)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetByParam)
	r.Get("/", h.HandleGetAll)
	r.Post("/update", h.HandleUpdate)
	r.Post("/update/", h.HandleUpdate)
	r.Post("/value", h.HandleGet)
	r.Post("/value/", h.HandleGet)

	return r
}

func (h *Handler) HandleUpdateByParam(res http.ResponseWriter, req *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(req, "metricType"),
		chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")

	res.Header().Set(headers.ContentType, "text/plain")

	m := models.Convert(metricType, metricName, metricValue)
	err := h.service.Validate(m)
	if err != nil {
		if errors.Is(err, service.ErrIDIsEmpty) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.service.Save(m)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGetByParam(res http.ResponseWriter, req *http.Request) {
	metricType, metricName := chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName")
	res.Header().Set(headers.ContentType, "text/plain")

	value, err := h.service.GetMetricValue(metricType, metricName)

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
	metrics, err := h.service.GetAll()
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
	m, err := models.ExtractBody(req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.service.Validate(*m)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	*m, err = h.service.Save(*m)
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
	m, err := models.ExtractBody(req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if m.MType != models.CounterType && m.MType != models.GaugeType {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	*m, err = h.service.Get((*m).MType, (*m).ID)
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
