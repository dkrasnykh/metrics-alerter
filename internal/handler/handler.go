package handler

import (
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
			{{range .Metrics}}<div>{{ .Type }} {{ .Name }} {{ .ValueInt64 }} {{ .ValueFloat64 }}</div>{{end}}
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

	r.Use(h.Logger)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdate)
	r.Get("/value/{metricType}/{metricName}", h.HandleGet)
	r.Get("/", h.HandleGetAll)

	return r
}

func (h *Handler) HandleUpdate(res http.ResponseWriter, req *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(req, "metricType"),
		chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")
	res.Header().Set(headers.ContentType, "text/plain")
	if metricName == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	err := h.service.Validate(metricType, metricValue)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.service.Save(metricType, metricName, metricValue)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) HandleGet(res http.ResponseWriter, req *http.Request) {
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
}

func (h *Handler) HandleGetAll(res http.ResponseWriter, req *http.Request) {
	type Item struct {
		Metrics []models.Metric
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
}
