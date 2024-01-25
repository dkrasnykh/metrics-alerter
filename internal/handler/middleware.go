package handler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/utils"
)

type CompressWriter struct {
	http.ResponseWriter
	Writer io.Writer
	bytes  []byte
}

func (w CompressWriter) Write(b []byte) (int, error) {
	w.bytes = b
	return w.Writer.Write(b)
}

func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(timestamp)
		logger.InfoRequest(r.Method, r.RequestURI, duration)
	})
}

func (h *Handler) GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(headers.ContentEncoding), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		oldBody := r.Body
		defer func(oldBody io.ReadCloser) {
			err := oldBody.Close()
			if err != nil {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(oldBody)
		zr, err := gzip.NewReader(oldBody)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = zr
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) GzipResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(headers.AcceptEncoding), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(gz)
		w.Header().Set(headers.ContentEncoding, "gzip")
		next.ServeHTTP(CompressWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func (h *Handler) Hash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(utils.HashHeader) != "" {
			expected := r.Header.Get(utils.HashHeader)
			buf, err := io.ReadAll(r.Body)
			utils.LogError(err)
			actual := utils.Hash(buf, []byte(h.key))
			if expected != actual {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(buf))
		}

		next.ServeHTTP(w, r)
		
		if h.key != "" {
			writer, ok := w.(CompressWriter)
			if ok {
				hash := utils.Hash(writer.bytes, []byte(h.key))
				w.Header().Set(utils.HashHeader, hash)
			}
		}
	})
}
