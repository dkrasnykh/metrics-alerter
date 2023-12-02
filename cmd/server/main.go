package main

import (
	"github.com/dkrasnykh/metrics-alerter/internal/handler"
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"log"
	"net/http"
)

func main() {
	r := storage.NewStorage()
	s := service.NewService(r)
	h := handler.NewHandler(s)

	mux := http.NewServeMux()
	mux.Handle("/update/", h)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err == nil {
		log.Fatalf("error occured while running http server: %s", err.Error())
	}
}
