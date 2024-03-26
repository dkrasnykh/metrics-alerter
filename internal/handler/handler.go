package handler

import (
	"context"
	"database/sql/driver"
	"html/template"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

//go:generate mockgen -source=handler.go -destination=mocks/mock.go
type Servicer interface {
	driver.Pinger
	Validate(m models.Metrics) error
	Save(ctx context.Context, m models.Metrics) (models.Metrics, error)
	GetMetricValue(ctx context.Context, mType, mName string) (string, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
	Get(ctx context.Context, mType, mName string) (models.Metrics, error)
	Load(ctx context.Context, metrics []models.Metrics) error
}

type Handler struct {
	service Servicer
	key     string
	tpl     *template.Template
}

func New(s Servicer, key string, t *template.Template) *Handler {
	return &Handler{
		service: s,
		key:     key,
		tpl:     t,
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(h.Hash)
	r.Use(h.GzipRequest)
	r.Use(h.GzipResponse)
	r.Use(h.Logging)

	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandleUpdateByParam)
	r.Post("/update/", h.HandleUpdate)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetByParam)
	r.Post("/value/", h.HandleGet)
	r.Get("/", h.HandleGetAll)
	r.Get("/ping", h.HandleGetPing)
	r.Post("/updates/", h.HandleUpdates)

	return r
}
