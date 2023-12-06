package handler

import (
	"github.com/dkrasnykh/metrics-alerter/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"html/template"
	"log"
	"net/http"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{s}
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
	} else {
		res.WriteHeader(http.StatusOK)
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
	const tpl = `
	<!DOCTYPE html>
	<html>
		<body>
			{{range .Items}}
				<h3>{{.Type}}</h3>
				<body>
					{{range .Values}}<div>{{ .Name }} {{ .Value }}</div>{{end}}
				</body>
			{{end}}
		</body>
	</html>`

	type Value struct {
		Name  string
		Value string
	}

	type Item struct {
		Type   string
		Values []Value
	}

	type Items struct {
		Items []Item
	}

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	t, err := template.New("webpage").Parse(tpl)
	check(err)

	values, err := h.service.GetAll()
	check(err)

	items := make([]Item, 0, len(values))
	for mtype, mname := range values {
		item := Item{Type: mtype, Values: make([]Value, 0)}
		for _, v := range mname {
			item.Values = append(item.Values, Value{Name: v[0], Value: v[1]})
		}
		items = append(items, item)
	}
	err = t.Execute(res, Items{Items: items})
	check(err)
}
