package handler

import (
	"html/template"
	"net/http"

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
			{{range .Counters}}<div>{{ .Name }} {{ .Value }}</div>{{end}}
			{{range .Gauges}}<div>{{ .Name }} {{ .Value }}</div>{{end}}
		</body>
	</html>`
)

var T *template.Template

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) InitRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdate)
	router.Get("/value/{metricType}/{metricName}", h.HandleGet)
	router.Get("/", h.HandleGetAll)

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
		Counters []models.Counter
		Gauges   []models.Gauge
	}
	counters, err := h.service.GetAllCounter()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	gauges, err := h.service.GetAllGauge()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = T.Execute(res, Item{Counters: counters, Gauges: gauges})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
