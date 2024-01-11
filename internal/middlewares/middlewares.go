package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/compress"
	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	Status int
	Size   int
}

type contextKey string

const userIDKey contextKey = "user_id"

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

func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := &auth.Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(auth.SecretKey), nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}
