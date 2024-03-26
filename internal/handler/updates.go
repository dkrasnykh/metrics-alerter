package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

func (h *Handler) HandleUpdates(res http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil || string(bytes) == `null` {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metrics := []models.Metrics{}
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, m := range metrics {
		err = h.service.Validate(m)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	err = h.service.Load(req.Context(), metrics)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
