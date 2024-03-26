package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dkrasnykh/metrics-alerter/internal/config"
	"github.com/dkrasnykh/metrics-alerter/internal/handler"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
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
	var err error
	r, err := storage.New(s.c)
	if err != nil {
		return err
	}
	v := service.New(r)
	tpl, err := template.New("webpage").Parse(handler.Tpl)
	if err != nil {
		return err
	}
	h := handler.New(v, s.c.Key, tpl)

	err = http.ListenAndServe(s.c.Address, h.InitRoutes())
	return err
}
