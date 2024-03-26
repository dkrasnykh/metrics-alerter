package handler

import (
	"bytes"
	"compress/gzip"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-http-utils/headers"

	"github.com/dkrasnykh/metrics-alerter/internal/hash"
)

type CompressWriter struct {
	http.ResponseWriter
	Writer io.Writer
	bytes  []byte
}

func (w CompressWriter) Write(b []byte) (int, error) {
	w.bytes = b
	return w.Writer.Write(w.bytes)
}

func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(timestamp)
		zap.L().Info("request",
			zap.String("method", r.Method),
			zap.String("URI", r.RequestURI),
			zap.Duration("duration", duration))
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
				zap.L().Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(oldBody)
		zr, err := gzip.NewReader(oldBody)
		if err != nil {
			zap.L().Error(err.Error())
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
			zap.L().Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				zap.L().Error(err.Error())
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
		if r.Header.Get(hash.Header) != "" {
			expected := r.Header.Get(hash.Header)
			buf, err := io.ReadAll(r.Body)
			if err != nil {
				zap.L().Error(err.Error())
			}
			actual := hash.Encode(buf, []byte(h.key))
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
				value := hash.Encode(writer.bytes, []byte(h.key))
				w.Header().Set(hash.Header, value)
			}
		}
	})
}
