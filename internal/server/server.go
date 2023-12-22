package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dkrasnykh/metrics-alerter/internal/handler"
	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

type Server struct {
	address string
	port    string
	router  *chi.Mux
}

func New(address string) *Server {
	return &Server{
		address: address,
		router:  chi.NewRouter(),
	}
}

func (s *Server) Run() error {
	r := storage.New()
	v := service.New(r)
	var err error
	handler.T, err = template.New("webpage").Parse(handler.Tpl)
	if err != nil {
		return err
	}
	l, err := logger.New()
	if err != nil {
		return err
	}
	h := handler.New(v, l)

	err = http.ListenAndServe(s.address, h.InitRoutes())
	if err != nil {
		return err
	}
	return nil
}
