package server

import (
	"fmt"
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

func NewServer(address, port string) *Server {
	return &Server{
		address: address,
		port:    port,
		router:  chi.NewRouter(),
	}
}

func (s *Server) Run() {
	r := storage.NewStorage()
	v := service.NewService(r)
	h := handler.NewHandler(v)

	err := http.ListenAndServe(fmt.Sprintf("%s:%s", s.address, s.port), h.InitRoutes())

	if err != nil {
		log.Fatal(err)
	}
}
