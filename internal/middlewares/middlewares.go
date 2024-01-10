package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/ilya-burinskiy/gophermart/internal/compress"
	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	Status int
	Size   int
}

func (lw *loggingResponseWriter) Write(bytes []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(bytes)
	lw.Size = size

	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(status int) {
	lw.ResponseWriter.WriteHeader(status)
	lw.Status = status
}

func LogResponse(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lw := loggingResponseWriter{ResponseWriter: w}
			h.ServeHTTP(&lw, r)
			logger.Info("response",
				zap.Int("status", lw.Status),
				zap.Int("size", lw.Size),
			)
		})
	}
}

func LogRequest(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lw := loggingResponseWriter{ResponseWriter: w}
			start := time.Now()
			h.ServeHTTP(&lw, r)
			duration := time.Since(start)
			logger.Info("got incoming http request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("duration", duration.String()),
			)
		})
	}
}

func GzipCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "gzip") {
			compressReader, err := compress.NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = compressReader
			defer compressReader.Close()
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			compressRw := compress.NewGzipWriter(w)
			w = compressRw
			defer compressRw.Close()
		}

		h.ServeHTTP(w, r)
	})
}
