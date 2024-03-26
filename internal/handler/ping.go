package handler

import "net/http"

func (h *Handler) HandleGetPing(res http.ResponseWriter, req *http.Request) {
	err := h.service.Ping(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
