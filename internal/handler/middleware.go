package handler

import (
	"io"
	"net/http"
	"time"
)

func (h *Handler) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Now().Sub(timestamp)
		h.logger.InfoRequest(r.Method, r.RequestURI, duration)
		if r.Response != nil {
			bytes, _ := io.ReadAll(r.Body)
			h.logger.InfoResponse(r.Response.StatusCode, len(bytes))
			_ = r.Body.Close()
		}
	})
}
