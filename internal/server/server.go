package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/handler"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/database"
	"github.com/dkrasnykh/metrics-alerter/internal/storage/memory"
)

type Server struct {
	c      *config.ServerConfig
	router *chi.Mux
}

func New(conf *config.ServerConfig) *Server {
	return &Server{
		c:      conf,
		router: chi.NewRouter(),
	}
}

func (s *Server) Run() error {
	r := memory.New(s.c.FileStoragePath, s.c.StoreInterval)
	if s.c.Restore {
		r.Restore()
	}
	v := service.New(r)
	database.Url = s.c.DatabaseDSN
	var err error
	handler.T, err = template.New("webpage").Parse(handler.Tpl)
	if err != nil {
		return err
	}
	h := handler.New(v)

	err = http.ListenAndServe(s.c.Address, h.InitRoutes())
	return err
}
