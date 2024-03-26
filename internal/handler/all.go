package handler

import (
	"net/http"

	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

const (
	Tpl = `
	<!DOCTYPE html>
	<html>
		<body>
			{{range .Metrics}}<div>{{ .MType }} {{ .ID }} {{ .Delta }} {{ .Value }}</div>{{end}}
		</body>
	</html>`
)

func (h *Handler) HandleGetAll(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(headers.ContentType, `text/html`)
	type Item struct {
		Metrics []models.Metrics
	}
	metrics, err := h.service.GetAll(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.tpl.Execute(res, Item{Metrics: metrics})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
