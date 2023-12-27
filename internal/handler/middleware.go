package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
)

type CompressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w CompressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(timestamp)
		h.logger.InfoRequest(r.Method, r.RequestURI, duration)
	})
}

func (h *Handler) GzipResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		oldBody := r.Body
		defer func(oldBody io.ReadCloser) {
			err := oldBody.Close()
			if err != nil {
				h.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(oldBody)
		zr, err := gzip.NewReader(oldBody)
		if err != nil {
			h.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = zr
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			h.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				h.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(gz)
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(CompressWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
