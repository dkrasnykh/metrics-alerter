package server

import (
	"github.com/dkrasnykh/metrics-alerter/internal/handler"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Server struct {
	address string
	port    string
	router  *chi.Mux
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		router:  chi.NewRouter(),
	}
}

func (s *Server) Run() {
	r := storage.NewStorage()
	v := service.NewService(r)
	h := handler.NewHandler(v)

	err := http.ListenAndServe(s.address, h.InitRoutes())

	if err != nil {
		log.Fatal(err)
	}
}
